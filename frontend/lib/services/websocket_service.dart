import 'dart:async';
import 'dart:convert';
import 'package:web_socket_channel/web_socket_channel.dart';
import '../core/constants/api_constants.dart';
import '../core/services/app_logger.dart';

class WebSocketService {
  WebSocketChannel? _channel;
  final StreamController<Map<String, dynamic>> _messageController =
      StreamController<Map<String, dynamic>>.broadcast();

  Stream<Map<String, dynamic>> get messages => _messageController.stream;

  void connect(String token) {
    if (_channel != null) return;

    // The backend authenticates via the FIRST WebSocket frame, not a query
    // param. Putting the JWT in the URL leaks it via logs/proxies/referrers
    // and is also rejected by the server ("first message must be auth").
    final channel = WebSocketChannel.connect(Uri.parse(ApiConstants.chatWs));
    _channel = channel;
    channel.sink.add(jsonEncode({'type': 'auth', 'token': token}));
    AppLogger.info('Call WS connecting');

    channel.stream.listen(
      (message) {
        try {
          final decoded = jsonDecode(message as String);
          if (decoded is Map<String, dynamic> &&
              !_messageController.isClosed) {
            _messageController.add(decoded);
          }
        } catch (e) {
          // Malformed server frame — skip it instead of killing the stream.
          AppLogger.warn('Call WS dropped malformed frame', e);
        }
      },
      onDone: () {
        AppLogger.info('Call WS closed');
        _channel = null;
      },
      onError: (error) {
        AppLogger.error('Call WS error', error);
        _channel = null;
      },
      cancelOnError: false,
    );
  }

  void disconnect() {
    AppLogger.info('Call WS disconnect requested');
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
