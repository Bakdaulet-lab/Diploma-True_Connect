import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_webrtc/flutter_webrtc.dart';
import '../../services/websocket_service.dart';

class CallScreen extends StatefulWidget {
  final Map<String, dynamic> matchData;
  final WebSocketService wsService;
  final bool isCaller;
  final Map<String, dynamic>? initialOffer;

  const CallScreen({
    super.key,
    required this.matchData,
    required this.wsService,
    required this.isCaller,
    this.initialOffer,
  });

  @override
  State<CallScreen> createState() => _CallScreenState();
}

class _CallScreenState extends State<CallScreen> {
  final RTCVideoRenderer _localRenderer = RTCVideoRenderer();
  final RTCVideoRenderer _remoteRenderer = RTCVideoRenderer();
  
  RTCPeerConnection? _peerConnection;
  MediaStream? _localStream;
  StreamSubscription<Map<String, dynamic>>? _wsSub;

  bool _isMicMuted = false;
  bool _isVideoTurnedOff = false;

  @override
  void initState() {
    super.initState();
    _initRenderers();
    _initWebRTC();
    _subscribeToSignaling();
  }

  void _subscribeToSignaling() {
    _wsSub = widget.wsService.messages.listen(
      (msg) {
        if (!mounted) return;
        try {
          final type = msg['type'] as String?;
          if (type == null || !type.startsWith('webrtc_')) return;

          final payload = msg['payload'];
          if (payload is! Map<String, dynamic>) return;

          final incomingMatchId = payload['match_id'] as String?;
          if (incomingMatchId != widget.matchData['id']) return;

          switch (type) {
            case 'webrtc_offer' when !widget.isCaller:
              _handleOffer(payload);
            case 'webrtc_answer' when widget.isCaller:
              _setRemoteDescription(payload);
            case 'webrtc_ice_candidate':
              _addCandidate(payload);
          }
        } catch (e, st) {
          debugPrint('[CallScreen] signaling error: $e\n$st');
        }
      },
      onError: (e) => debugPrint('[CallScreen] WS stream error: $e'),
      cancelOnError: false,
    );
  }

  Future<void> _initRenderers() async {
    await _localRenderer.initialize();
    await _remoteRenderer.initialize();
  }

  Future<void> _initWebRTC() async {
    // 1. Get local media
    final Map<String, dynamic> mediaConstraints = {
      'audio': true,
      'video': {
        'mandatory': {
          'minWidth': '640',
          'minHeight': '480',
          'minFrameRate': '30',
        },
        'facingMode': 'user',
        'optional': [],
      }
    };

    try {
      _localStream = await navigator.mediaDevices.getUserMedia(mediaConstraints);
      _localRenderer.srcObject = _localStream;
    } catch (e) {
      debugPrint('Error getting user media: $e');
      return;
    }

    // 2. Create peer connection
    final Map<String, dynamic> configuration = {
      'iceServers': [
        {'urls': 'stun:stun.l.google.com:19302'},
      ]
    };

    _peerConnection = await createPeerConnection(configuration);

    // 3. Add local stream to peer connection
    _localStream?.getTracks().forEach((track) {
      _peerConnection?.addTrack(track, _localStream!);
    });

    // 4. Handle incoming remote stream
    _peerConnection?.onAddStream = (MediaStream stream) {
      setState(() {
        _remoteRenderer.srcObject = stream;
      });
    };

    // 5. Handle ICE candidates
    _peerConnection?.onIceCandidate = (RTCIceCandidate candidate) {
      widget.wsService.sendWebRTCSignal('webrtc_ice_candidate', widget.matchData['id'], {
        'candidate': candidate.candidate,
        'sdpMid': candidate.sdpMid,
        'sdpMLineIndex': candidate.sdpMLineIndex,
      });
    };

    // 6. Connect!
    if (widget.isCaller) {
      _createOffer();
    } else if (widget.initialOffer != null) {
      _handleOffer(widget.initialOffer!);
    }
  }

  Future<void> _createOffer() async {
    final RTCSessionDescription description = await _peerConnection!.createOffer();
    await _peerConnection!.setLocalDescription(description);

    widget.wsService.sendWebRTCSignal('webrtc_offer', widget.matchData['id'], {
      'sdp': description.sdp,
    });
  }

  Future<void> _handleOffer(Map<String, dynamic> offerPayload) async {
    final RTCSessionDescription description = RTCSessionDescription(
      offerPayload['sdp'],
      'offer',
    );
    await _peerConnection?.setRemoteDescription(description);

    final RTCSessionDescription answer = await _peerConnection!.createAnswer();
    await _peerConnection!.setLocalDescription(answer);

    widget.wsService.sendWebRTCSignal('webrtc_answer', widget.matchData['id'], {
      'sdp': answer.sdp,
    });
  }

  Future<void> _setRemoteDescription(Map<String, dynamic> answerPayload) async {
    final RTCSessionDescription description = RTCSessionDescription(
      answerPayload['sdp'],
      'answer',
    );
    await _peerConnection?.setRemoteDescription(description);
  }

  Future<void> _addCandidate(Map<String, dynamic> payload) async {
    if (payload['candidate'] != null) {
      final RTCIceCandidate candidate = RTCIceCandidate(
        payload['candidate'],
        payload['sdpMid'],
        payload['sdpMLineIndex'],
      );
      await _peerConnection?.addCandidate(candidate);
    }
  }

  void _toggleMic() {
    if (_localStream != null) {
      final bool muted = !_isMicMuted;
      _localStream!.getAudioTracks().forEach((track) {
        track.enabled = !muted;
      });
      setState(() {
        _isMicMuted = muted;
      });
    }
  }

  void _toggleVideo() {
    if (_localStream != null) {
      final bool videoOff = !_isVideoTurnedOff;
      _localStream!.getVideoTracks().forEach((track) {
        track.enabled = !videoOff;
      });
      setState(() {
        _isVideoTurnedOff = videoOff;
      });
    }
  }

  void _endCall() {
    _localStream?.getTracks().forEach((track) => track.stop());
    _peerConnection?.close();
    Navigator.of(context).pop();
  }

  @override
  void dispose() {
    _wsSub?.cancel();
    _localRenderer.dispose();
    _remoteRenderer.dispose();
    _localStream?.getTracks().forEach((track) => track.stop());
    _localStream?.dispose();
    _peerConnection?.dispose();
    widget.wsService.disconnect();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.black,
      body: SafeArea(
        child: Stack(
          children: [
            // Remote Video Fullscreen
            Positioned.fill(
              child: _remoteRenderer.srcObject != null 
                  ? RTCVideoView(
                      _remoteRenderer,
                      objectFit: RTCVideoViewObjectFit.RTCVideoViewObjectFitCover,
                    )
                  : const Center(
                      child: CircularProgressIndicator(color: Colors.white),
                    ),
            ),
            
            // Local Video Floating
            Positioned(
              top: 20,
              right: 20,
              width: 120,
              height: 160,
              child: ClipRRect(
                borderRadius: BorderRadius.circular(16),
                child: Container(
                  color: Colors.black54,
                  child: RTCVideoView(
                    _localRenderer,
                    mirror: true,
                    objectFit: RTCVideoViewObjectFit.RTCVideoViewObjectFitCover,
                  ),
                ),
              ),
            ),

            // Controls Bottom Bar
            Positioned(
              bottom: 40,
              left: 0,
              right: 0,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  FloatingActionButton(
                    heroTag: 'mic',
                    backgroundColor: _isMicMuted ? Colors.red : Colors.white,
                    onPressed: _toggleMic,
                    child: Icon(
                      _isMicMuted ? Icons.mic_off : Icons.mic,
                      color: _isMicMuted ? Colors.white : Colors.black,
                    ),
                  ),
                  FloatingActionButton(
                    heroTag: 'end',
                    backgroundColor: Colors.red,
                    onPressed: _endCall,
                    child: const Icon(Icons.call_end, color: Colors.white),
                  ),
                  FloatingActionButton(
                    heroTag: 'video',
                    backgroundColor: _isVideoTurnedOff ? Colors.red : Colors.white,
                    onPressed: _toggleVideo,
                    child: Icon(
                      _isVideoTurnedOff ? Icons.videocam_off : Icons.videocam,
                      color: _isVideoTurnedOff ? Colors.white : Colors.black,
                    ),
                  ),
                ],
              ),
            ),
            
            // Blind Date overlay logic
            if (_isVideoTurnedOff)
              Positioned.fill(
                child: Container(
                  color: Colors.black87,
                  child: const Center(
                    child: Text(
                      'Video is paused\n(Blind Date Mode)',
                      textAlign: TextAlign.center,
                      style: TextStyle(color: Colors.white70, fontSize: 18),
                    ),
                  ),
                ),
              ),
            
            Positioned(
              top: 20,
              left: 20,
              child: IconButton(
                icon: const Icon(Icons.arrow_back, color: Colors.white),
                onPressed: _endCall,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
