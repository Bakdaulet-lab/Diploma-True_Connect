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
  final String _currentUserId;
  WebSocketChannel? _channel;
  StreamSubscription<dynamic>? _sub;
  Timer? _typingTimer;
  Timer? _reconnectTimer;
  int _reconnectAttempts = 0;
  bool _disposed = false; // set in dispose(), cannot be final
  static const _maxReconnectAttempts = 5;
  static const _storage = FlutterSecureStorage();

  final _eventBus = StreamController<ChatEvent>.broadcast();
  Stream<ChatEvent> get events => _eventBus.stream;

  ChatNotifier(this._matchId, this._dio, this._currentUserId)
      : super(const ChatState()) {
    _connect();
  }

  Future<void> _connect() async {
    if (_disposed) return;
    final token = await _storage.read(key: 'access_token') ?? '';
    final uri = Uri.parse(ApiConstants.chatWs);
    _channel = WebSocketChannel.connect(uri);

    // Backend requires first message to be auth handshake before anything else.
    _channel!.sink.add(jsonEncode({'type': 'auth', 'token': token}));

    _sub = _channel!.stream.listen(
      _onRaw,
      onDone: _onDisconnected,
      onError: (_) => _onDisconnected(),
    );
  }

  void _onDisconnected() {
    if (_disposed) return;
    state = state.copyWith(isConnected: false);
    _scheduleReconnect();
  }

  void _scheduleReconnect() {
    if (_disposed || _reconnectAttempts >= _maxReconnectAttempts) return;
    // Exponential backoff: 1s, 2s, 4s, 8s, 16s
    final delay = Duration(seconds: 1 << _reconnectAttempts);
    _reconnectAttempts++;
    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(delay, () async {
      await _sub?.cancel();
      await _connect();
    });
  }

  // Call after successful auth_ok to reset the backoff counter.
  void _onConnected() {
    _reconnectAttempts = 0;
    state = state.copyWith(isConnected: true);
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

      // Handle auth handshake response — load history once authenticated.
      if (json['type'] == 'auth_ok') {
        _onConnected();
        return;
      }

      final event = ChatEvent.fromJson(json);
      _eventBus.add(event);

      switch (event.type) {
        case ChatEventType.chatMsg:
        case ChatEventType.contentWarning:
          if (event.payload['match_id'] == _matchId) {
            // Replace optimistic temp message with confirmed one if content matches.
            final msgs = List<Map<String, dynamic>>.from(state.messages);
            final content = event.payload['content'] as String?;
            final senderId = event.payload['sender_id'] as String?;
            final tempIdx = senderId == _currentUserId
                ? msgs.lastIndexWhere((m) =>
                    (m['id'] as String? ?? '').startsWith('temp_') &&
                    m['content'] == content)
                : -1;
            if (tempIdx != -1) {
              msgs[tempIdx] = Map<String, dynamic>.from(event.payload);
            } else {
              msgs.add(Map<String, dynamic>.from(event.payload));
            }
            state = state.copyWith(messages: msgs, isOtherTyping: false);
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
    if (!state.isConnected) return;
    _channel?.sink.add(jsonEncode({
      'type': 'chat_msg',
      'payload': {'match_id': _matchId, 'content': content},
    }));
    // Optimistic local insert using actual user ID so bubbles render correctly.
    state = state.copyWith(messages: [
      ...state.messages,
      {
        'id': 'temp_${DateTime.now().millisecondsSinceEpoch}',
        'sender_id': _currentUserId,
        'match_id': _matchId,
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
    _disposed = true;
    _reconnectTimer?.cancel();
    _typingTimer?.cancel();
    _sub?.cancel();
    _channel?.sink.close();
    _eventBus.close();
    super.dispose();
  }
}

final chatNotifierProvider =
    StateNotifierProvider.family<ChatNotifier, ChatState, String>(
  (ref, matchId) {
    final userId = ref.watch(authStateProvider).valueOrNull?.id ?? '';
    return ChatNotifier(matchId, ref.watch(dioClientProvider).dio, userId);
  },
);
