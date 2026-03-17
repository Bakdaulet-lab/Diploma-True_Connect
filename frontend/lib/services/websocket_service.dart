import 'dart:async';
import 'dart:convert';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import '../core/constants/api_constants.dart';
import '../core/storage/token_storage.dart';
import '../core/network/dio_client.dart';
import '../models/match.dart';

final webSocketServiceProvider = Provider<WebSocketService>((ref) {
  return WebSocketService(ref.read(tokenStorageProvider));
});

class WebSocketService {
  final TokenStorage _tokenStorage;
  WebSocketChannel? _channel;
  final _messageController = StreamController<Message>.broadcast();
  final _connectionController = StreamController<bool>.broadcast();
  bool _isConnected = false;
  Timer? _reconnectTimer;
  Timer? _pingTimer;

  WebSocketService(this._tokenStorage);

  Stream<Message> get messageStream => _messageController.stream;
  Stream<bool> get connectionStream => _connectionController.stream;
  bool get isConnected => _isConnected;

  Future<void> connect() async {
    if (_isConnected) return;

    try {
      final token = await _tokenStorage.getAccessToken();
      if (token == null) return;

      _channel = WebSocketChannel.connect(
        Uri.parse(ApiConstants.ws),
      );

      // Send auth token as first message
      _channel!.sink.add(jsonEncode({
        'type': 'auth',
        'token': token,
      }));

      _isConnected = true;
      _connectionController.add(true);

      _channel!.stream.listen(
        (data) {
          try {
            final json = jsonDecode(data as String);
            if (json['type'] == 'message') {
              final message = Message.fromJson(json['data'] ?? json);
              _messageController.add(message);
            }
          } catch (_) {}
        },
        onDone: () {
          _isConnected = false;
          _connectionController.add(false);
          _scheduleReconnect();
        },
        onError: (error) {
          _isConnected = false;
          _connectionController.add(false);
          _scheduleReconnect();
        },
      );

      _startPing();
    } catch (_) {
      _isConnected = false;
      _connectionController.add(false);
      _scheduleReconnect();
    }
  }

  void sendMessage({
    required String matchId,
    required String content,
  }) {
    if (!_isConnected || _channel == null) return;

    _channel!.sink.add(jsonEncode({
      'type': 'message',
      'match_id': matchId,
      'content': content,
    }));
  }

  void sendTyping(String matchId) {
    if (!_isConnected || _channel == null) return;

    _channel!.sink.add(jsonEncode({
      'type': 'typing',
      'match_id': matchId,
    }));
  }

  void _startPing() {
    _pingTimer?.cancel();
    _pingTimer = Timer.periodic(const Duration(seconds: 30), (_) {
      if (_isConnected && _channel != null) {
        _channel!.sink.add(jsonEncode({'type': 'ping'}));
      }
    });
  }

  void _scheduleReconnect() {
    _reconnectTimer?.cancel();
    _reconnectTimer = Timer(const Duration(seconds: 5), () {
      connect();
    });
  }

  void disconnect() {
    _pingTimer?.cancel();
    _reconnectTimer?.cancel();
    _channel?.sink.close();
    _isConnected = false;
    _connectionController.add(false);
  }

  void dispose() {
    disconnect();
    _messageController.close();
    _connectionController.close();
  }
}
