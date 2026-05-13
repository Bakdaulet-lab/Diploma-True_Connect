import 'dart:async';
import 'dart:convert';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import '../core/constants/api_constants.dart';
import 'auth_provider.dart';

// ─── Chat Event ───────────────────────────────────────────────────────────────

enum ChatEventType {
  chatMsg,
  contentWarning,
  contentBlocked,
  mahramChatMsg,
  mahramRoomInvite,
  typing,
  read,
  unknown,
}

class ChatEvent {
  final ChatEventType type;
  final Map<String, dynamic> payload;

  const ChatEvent({required this.type, required this.payload});

  static ChatEventType _parseType(String t) {
    switch (t) {
      case 'chat_msg':
        return ChatEventType.chatMsg;
      case 'content_warning':
        return ChatEventType.contentWarning;
      case 'content_blocked':
        return ChatEventType.contentBlocked;
      case 'mahram_chat_msg':
        return ChatEventType.mahramChatMsg;
      case 'mahram_room_invite':
        return ChatEventType.mahramRoomInvite;
      case 'typing':
        return ChatEventType.typing;
      case 'read':
        return ChatEventType.read;
      default:
        return ChatEventType.unknown;
    }
  }

  factory ChatEvent.fromJson(Map<String, dynamic> json) => ChatEvent(
        type: _parseType(json['type'] as String? ?? ''),
        payload: json['payload'] as Map<String, dynamic>? ?? {},
      );
}

// ─── Chat State ───────────────────────────────────────────────────────────────

class ChatState {
  final List<Map<String, dynamic>> messages;
  final bool isConnected;
  final bool isOtherTyping;

  const ChatState({
    this.messages = const [],
    this.isConnected = false,
    this.isOtherTyping = false,
  });

  ChatState copyWith({
    List<Map<String, dynamic>>? messages,
    bool? isConnected,
    bool? isOtherTyping,
  }) =>
      ChatState(
        messages: messages ?? this.messages,
        isConnected: isConnected ?? this.isConnected,
        isOtherTyping: isOtherTyping ?? this.isOtherTyping,
      );
}

// ─── Chat Notifier ────────────────────────────────────────────────────────────

class ChatNotifier extends StateNotifier<ChatState> {
  final String _matchId;
  final Dio _dio;
  WebSocketChannel? _channel;
  StreamSubscription<dynamic>? _sub;
  Timer? _typingTimer;
  static const _storage = FlutterSecureStorage();

  final _eventBus = StreamController<ChatEvent>.broadcast();
  Stream<ChatEvent> get events => _eventBus.stream;

  ChatNotifier(this._matchId, this._dio) : super(const ChatState()) {
    _connect();
  }

  Future<void> _connect() async {
    final token = await _storage.read(key: 'access_token') ?? '';
    final uri = Uri.parse('${ApiConstants.chatWs}?token=$token');
    _channel = WebSocketChannel.connect(uri);
    state = state.copyWith(isConnected: true);
    _sub = _channel!.stream.listen(
      _onRaw,
      onDone: () => state = state.copyWith(isConnected: false),
      onError: (_) => state = state.copyWith(isConnected: false),
    );
    _loadHistory();
  }

  Future<void> _loadHistory() async {
    try {
      final resp = await _dio.get(
        ApiConstants.matchMessages(_matchId),
        queryParameters: {'limit': 50},
      );
      final raw = resp.data is Map ? resp.data['data'] ?? resp.data : resp.data;
      if (raw is! List) return;
      final history = raw
          .whereType<Map>()
          .map((e) => Map<String, dynamic>.from(e))
          .toList();
      if (history.isEmpty) return;
      // Prepend history; deduplicate against any live messages already received
      final liveIds = {for (final m in state.messages) m['id'] as String? ?? ''};
      final deduped = history.where((m) => !liveIds.contains(m['id'])).toList();
      state = state.copyWith(messages: [...deduped, ...state.messages]);
    } catch (_) {
      // History load is best-effort — don't disrupt the live chat.
    }
  }

  void _onRaw(dynamic raw) {
    try {
      final json = jsonDecode(raw as String) as Map<String, dynamic>;
      final event = ChatEvent.fromJson(json);
      _eventBus.add(event);

      switch (event.type) {
        case ChatEventType.chatMsg:
        case ChatEventType.contentWarning:
          if (event.payload['match_id'] == _matchId) {
            state = state.copyWith(
              messages: [...state.messages, event.payload],
              isOtherTyping: false,
            );
          }
        case ChatEventType.typing:
          if (event.payload['match_id'] == _matchId) {
            state = state.copyWith(isOtherTyping: true);
            _typingTimer?.cancel();
            _typingTimer = Timer(const Duration(seconds: 3), () {
              if (mounted) state = state.copyWith(isOtherTyping: false);
            });
          }
        default:
          break;
      }
    } catch (_) {}
  }

  void send(String content) {
    _channel?.sink.add(jsonEncode({
      'type': 'chat_msg',
      'payload': {'match_id': _matchId, 'content': content},
    }));
    // Optimistic local insert
    state = state.copyWith(messages: [
      ...state.messages,
      {
        'id': 'temp_${DateTime.now().millisecondsSinceEpoch}',
        'sender_id': 'me',
        'content': content,
        'is_read': false,
        'created_at': DateTime.now().toIso8601String(),
      },
    ]);
  }

  void sendTyping() {
    _channel?.sink.add(jsonEncode({
      'type': 'typing',
      'payload': {'match_id': _matchId},
    }));
  }

  void sendReadReceipt(String messageId) {
    _channel?.sink.add(jsonEncode({
      'type': 'read',
      'payload': {'message_id': messageId},
    }));
  }

  @override
  void dispose() {
    _typingTimer?.cancel();
    _sub?.cancel();
    _channel?.sink.close();
    _eventBus.close();
    super.dispose();
  }
}

final chatNotifierProvider =
    StateNotifierProvider.family<ChatNotifier, ChatState, String>(
  (ref, matchId) => ChatNotifier(matchId, ref.watch(dioClientProvider).dio),
);
