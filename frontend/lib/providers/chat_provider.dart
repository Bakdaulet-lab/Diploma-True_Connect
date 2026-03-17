import 'dart:async';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/models.dart';
import '../services/api_service.dart';
import '../services/websocket_service.dart';

class ChatState {
  final Map<String, List<Message>> messagesByMatch;
  final bool isLoading;
  final String? error;
  final bool isConnected;

  const ChatState({
    this.messagesByMatch = const {},
    this.isLoading = false,
    this.error,
    this.isConnected = false,
  });

  ChatState copyWith({
    Map<String, List<Message>>? messagesByMatch,
    bool? isLoading,
    String? error,
    bool? isConnected,
  }) =>
      ChatState(
        messagesByMatch: messagesByMatch ?? this.messagesByMatch,
        isLoading: isLoading ?? this.isLoading,
        error: error,
        isConnected: isConnected ?? this.isConnected,
      );

  List<Message> getMessages(String matchId) =>
      messagesByMatch[matchId] ?? [];
}

class ChatNotifier extends StateNotifier<ChatState> {
  final ApiService _apiService;
  final WebSocketService _wsService;
  StreamSubscription<Message>? _messageSubscription;
  StreamSubscription<bool>? _connectionSubscription;

  ChatNotifier(this._apiService, this._wsService) : super(const ChatState()) {
    _listenToWebSocket();
  }

  void _listenToWebSocket() {
    _messageSubscription = _wsService.messageStream.listen((message) {
      final matchMessages = state.getMessages(message.matchId);
      final updatedMessages = [...matchMessages, message];
      state = state.copyWith(
        messagesByMatch: {
          ...state.messagesByMatch,
          message.matchId: updatedMessages,
        },
      );
    });

    _connectionSubscription = _wsService.connectionStream.listen((connected) {
      state = state.copyWith(isConnected: connected);
    });
  }

  Future<void> connect() async {
    await _wsService.connect();
  }

  Future<void> loadMessages(String matchId) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final messages = await _apiService.getMessages(matchId);
      state = state.copyWith(
        messagesByMatch: {
          ...state.messagesByMatch,
          matchId: messages,
        },
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  void sendMessage({
    required String matchId,
    required String content,
  }) {
    _wsService.sendMessage(matchId: matchId, content: content);
  }

  void sendTyping(String matchId) {
    _wsService.sendTyping(matchId);
  }

  void disconnect() {
    _wsService.disconnect();
  }

  @override
  void dispose() {
    _messageSubscription?.cancel();
    _connectionSubscription?.cancel();
    _wsService.dispose();
    super.dispose();
  }
}

final chatProvider = StateNotifierProvider<ChatNotifier, ChatState>((ref) {
  return ChatNotifier(
    ref.read(apiServiceProvider),
    ref.read(webSocketServiceProvider),
  );
});
