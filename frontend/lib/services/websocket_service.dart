import 'dart:async';
import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';

class WebSocketService {
  WebSocketChannel? _channel;
  final StreamController<Map<String, dynamic>> _messageController =
      StreamController<Map<String, dynamic>>.broadcast();

  Stream<Map<String, dynamic>> get messages => _messageController.stream;

  void connect(String token) {
    if (_channel != null) return;

    final uri = Uri.parse('ws://localhost:8080/v1/chat/ws?token=$token');
    _channel = WebSocketChannel.connect(uri);

    _channel!.stream.listen(
      (message) {
        final decoded = jsonDecode(message);
        _messageController.add(decoded);
      },
      onDone: () {
        _channel = null;
        // Could implement reconnection logic here
      },
      onError: (error) {
        _channel = null;
        // Handle error
      },
    );
  }

  void disconnect() {
    _channel?.sink.close();
    _channel = null;
  }

  void sendMessage(String matchId, String content) {
    if (_channel == null) return;
    _channel!.sink.add(jsonEncode({
      'type': 'chat_msg',
      'payload': {
        'match_id': matchId,
        'content': content,
      }
    }));
  }

  void sendTyping(String matchId) {
    if (_channel == null) return;
    _channel!.sink.add(jsonEncode({
      'type': 'typing',
      'payload': {
        'match_id': matchId,
      }
    }));
  }

  void sendReadReceipt(String messageId) {
    if (_channel == null) return;
    _channel!.sink.add(jsonEncode({
      'type': 'read',
      'payload': {
        'message_id': messageId,
      }
    }));
  }

  void sendWebRTCSignal(String type, String matchId, Map<String, dynamic> data) {
    if (_channel == null) return;
    _channel!.sink.add(jsonEncode({
      'type': type, // 'webrtc_offer', 'webrtc_answer', 'webrtc_ice_candidate'
      'payload': {
        'match_id': matchId,
        'sdp': data['sdp'],
        'candidate': data['candidate'],
        'sdpMid': data['sdpMid'],
        'sdpMLineIndex': data['sdpMLineIndex'],
      }
    }));
  }
}
