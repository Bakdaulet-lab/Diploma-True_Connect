import 'dart:async';
import 'package:flutter/material.dart';
import '../../services/websocket_service.dart';
import 'call_screen.dart';

class ChatScreen extends StatefulWidget {
  final Map<String, dynamic> matchData;

  const ChatScreen({super.key, required this.matchData});

  @override
  State<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends State<ChatScreen> {
  final WebSocketService _wsService = WebSocketService();
  final TextEditingController _msgController = TextEditingController();
  final ScrollController _scrollController = ScrollController();
  
  List<Map<String, dynamic>> _messages = [];
  bool _isOtherTyping = false;
  Timer? _typingTimer;

  @override
  void initState() {
    super.initState();
    // Connect passing a dummy token (replace with actual auth token)
    _wsService.connect('dummy_token');

    _wsService.messages.listen((msg) {
      if (mounted) {
        setState(() {
          if (msg['type'] == 'typing' && msg['payload']['match_id'] == widget.matchData['id']) {
            _isOtherTyping = true;
            _typingTimer?.cancel();
            _typingTimer = Timer(const Duration(seconds: 3), () {
              if (mounted) setState(() => _isOtherTyping = false);
            });
          } else if (msg['type'] == 'chat_msg') {
            _messages.add(msg['payload']);
            _isOtherTyping = false;
            _scrollToBottom();
            
            // Send read receipt back
            if (msg['payload']['id'] != null) {
              _wsService.sendReadReceipt(msg['payload']['id']);
            }
          } else if (msg['type'] == 'read') {
            // Update UI to show message was read
            final msgId = msg['payload']['message_id'];
            final idx = _messages.indexWhere((m) => m['id'] == msgId);
            if (idx != -1) {
              _messages[idx]['is_read'] = true;
            }
          } else if (msg['type'] == 'webrtc_offer') {
            Navigator.push(context, MaterialPageRoute(
              builder: (_) => CallScreen(
                matchData: widget.matchData,
                wsService: _wsService,
                isCaller: false,
                initialOffer: msg['payload'],
              ),
            ));
          }
        });
      }
    });

    // Populate some initial dummy messages for UI demonstration
    _messages = [
      {
        'id': 'msg_1',
        'sender_id': widget.matchData['id'],
        'content': 'Hey! Thanks for matching. Love your prompts!',
        'is_read': true,
        'created_at': DateTime.now().subtract(const Duration(minutes: 5)).toIso8601String(),
      },
      {
        'id': 'msg_2',
        'sender_id': 'me',
        'content': 'Of course! Central Park walks are the best.',
        'is_read': true, // Simulated read receipt
        'created_at': DateTime.now().subtract(const Duration(minutes: 2)).toIso8601String(),
      }
    ];
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollController.hasClients) {
        _scrollController.animateTo(
          _scrollController.position.maxScrollExtent,
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeOut,
        );
      }
    });
  }

  void _sendMessage() {
    if (_msgController.text.trim().isEmpty) return;

    final content = _msgController.text.trim();
    _wsService.sendMessage(widget.matchData['id'], content);

    setState(() {
      _messages.add({
        'id': 'temp_${DateTime.now().millisecondsSinceEpoch}',
        'sender_id': 'me',
        'content': content,
        'is_read': false, // Initially unread
        'created_at': DateTime.now().toIso8601String(),
      });
    });
    
    _msgController.clear();
    _scrollToBottom();
  }

  void _onTyping(String text) {
    if (text.isNotEmpty) {
      _wsService.sendTyping(widget.matchData['id']);
    }
  }

  @override
  void dispose() {
    _wsService.disconnect();
    _msgController.dispose();
    _scrollController.dispose();
    _typingTimer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Row(
          children: [
            CircleAvatar(
              backgroundImage: NetworkImage(widget.matchData['imageUrl']),
              radius: 18,
            ),
            const SizedBox(width: 12),
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(widget.matchData['name'], style: const TextStyle(fontSize: 16)),
                if (_isOtherTyping)
                  const Text('typing...', style: TextStyle(fontSize: 12, color: Colors.green)),
              ],
            ),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.videocam),
            onPressed: () {
              Navigator.push(context, MaterialPageRoute(
                builder: (_) => CallScreen(
                  matchData: widget.matchData,
                  wsService: _wsService,
                  isCaller: true,
                ),
              ));
            },
          ),
        ],
      ),
      body: Column(
        children: [
          Expanded(
            child: ListView.builder(
              controller: _scrollController,
              padding: const EdgeInsets.all(16),
              itemCount: _messages.length,
              itemBuilder: (context, index) {
                final message = _messages[index];
                final isMe = message['sender_id'] == 'me';

                return Align(
                  alignment: isMe ? Alignment.centerRight : Alignment.centerLeft,
                  child: Container(
                    margin: const EdgeInsets.only(bottom: 8),
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                    decoration: BoxDecoration(
                      color: isMe ? Theme.of(context).primaryColor : Colors.grey[200],
                      borderRadius: BorderRadius.circular(20).copyWith(
                        bottomRight: isMe ? const Radius.circular(0) : const Radius.circular(20),
                        bottomLeft: !isMe ? const Radius.circular(0) : const Radius.circular(20),
                      ),
                    ),
                    child: Column(
                      crossAxisAlignment: isMe ? CrossAxisAlignment.end : CrossAxisAlignment.start,
                      children: [
                        Text(
                          message['content'],
                          style: TextStyle(
                            color: isMe ? Colors.white : Colors.black87,
                            fontSize: 15,
                          ),
                        ),
                        if (isMe) ...[
                          const SizedBox(height: 4),
                          Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Text(
                                _formatTime(message['created_at']),
                                style: const TextStyle(fontSize: 10, color: Colors.white70),
                              ),
                              const SizedBox(width: 4),
                              Icon(
                                message['is_read'] == true ? Icons.done_all : Icons.check,
                                size: 14,
                                color: message['is_read'] == true ? Colors.blue[200] : Colors.white70,
                              ),
                            ],
                          ),
                        ]
                      ],
                    ),
                  ),
                );
              },
            ),
          ),
          SafeArea(
            child: Padding(
              padding: const EdgeInsets.all(8.0),
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      controller: _msgController,
                      onChanged: _onTyping,
                      decoration: InputDecoration(
                        hintText: 'Type a message...',
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(24),
                          borderSide: BorderSide.none,
                        ),
                        filled: true,
                        fillColor: Colors.grey[200],
                        contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  CircleAvatar(
                    backgroundColor: Theme.of(context).primaryColor,
                    child: IconButton(
                      icon: const Icon(Icons.send, color: Colors.white),
                      onPressed: _sendMessage,
                    ),
                  )
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  String _formatTime(String isoTime) {
    if (isoTime.isEmpty) return '';
    try {
      final dt = DateTime.parse(isoTime).toLocal();
      final hr = dt.hour.toString().padLeft(2, '0');
      final mn = dt.minute.toString().padLeft(2, '0');
      return '$hr:$mn';
    } catch (_) {
      return '';
    }
  }
}