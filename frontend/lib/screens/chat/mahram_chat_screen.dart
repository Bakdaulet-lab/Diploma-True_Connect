import 'dart:async';
import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import '../../core/constants/api_constants.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../widgets/halal_pattern_painter.dart';

class MahramChatScreen extends ConsumerStatefulWidget {
  final String roomId;
  const MahramChatScreen({super.key, required this.roomId});

  @override
  ConsumerState<MahramChatScreen> createState() => _MahramChatScreenState();
}

class _MahramChatScreenState extends ConsumerState<MahramChatScreen> {
  final _msgCtr = TextEditingController();
  final _scrollCtr = ScrollController();
  static const _storage = FlutterSecureStorage();

  WebSocketChannel? _channel;
  StreamSubscription<dynamic>? _sub;
  bool _mahramJoined = false;

  final List<Map<String, dynamic>> _messages = [];

  final String _myId = 'me';

  @override
  void initState() {
    super.initState();
    _connect();
  }

  Future<void> _connect() async {
    final token = await _storage.read(key: 'access_token') ?? '';
    final uri = Uri.parse('${ApiConstants.chatWs}?token=$token');
    _channel = WebSocketChannel.connect(uri);

    _sub = _channel!.stream.listen((raw) {
      try {
        final json = jsonDecode(raw as String) as Map<String, dynamic>;
        if (json['type'] == 'mahram_chat_msg' &&
            json['payload']?['room_id'] == widget.roomId) {
          if (mounted) {
            setState(() {
              _messages.add(json['payload'] as Map<String, dynamic>);
              _mahramJoined = true;
            });
            _scrollToBottom();
          }
        } else if (json['type'] == 'mahram_room_invite') {
          if (mounted) setState(() => _mahramJoined = true);
        }
      } catch (_) {}
    }, onDone: () {});
  }

  void _send() {
    final text = _msgCtr.text.trim();
    if (text.isEmpty) return;
    _channel?.sink.add(jsonEncode({
      'type': 'mahram_chat_msg',
      'payload': {'room_id': widget.roomId, 'content': text},
    }));
    setState(() {
      _messages.add({
        'id': 'temp_${DateTime.now().millisecondsSinceEpoch}',
        'sender_id': _myId,
        'sender_role': 'woman',
        'content': text,
        'created_at': DateTime.now().toIso8601String(),
      });
    });
    _msgCtr.clear();
    _scrollToBottom();
  }

  void _scrollToBottom() {
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (_scrollCtr.hasClients) {
        _scrollCtr.animateTo(
          _scrollCtr.position.maxScrollExtent,
          duration: const Duration(milliseconds: 300),
          curve: Curves.easeOut,
        );
      }
    });
  }

  @override
  void dispose() {
    _sub?.cancel();
    _channel?.sink.close();
    _msgCtr.dispose();
    _scrollCtr.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: _buildAppBar(),
      body: Column(
        children: [
          // Mahram not yet joined banner
          if (!_mahramJoined)
            Container(
              color: AppColors.secondaryLight,
              padding: const EdgeInsets.symmetric(
                  horizontal: AppSpacing.md, vertical: 8),
              child: Row(
                children: [
                  const Icon(Icons.group_add,
                      size: 16, color: AppColors.secondary),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      'Махрам әлі қосылмаған. Telegram арқылы шақыру жіберілді.',
                      style: GoogleFonts.nunito(
                          fontSize: 12, color: AppColors.secondary),
                    ),
                  ),
                ],
              ),
            ),

          // Messages
          Expanded(
            child: _messages.isEmpty
                ? _buildEmptyState()
                : ListView.builder(
                    controller: _scrollCtr,
                    padding: const EdgeInsets.fromLTRB(
                        AppSpacing.md, AppSpacing.md, AppSpacing.md, 8),
                    itemCount: _messages.length,
                    itemBuilder: (_, i) =>
                        _MahramBubble(msg: _messages[i], myId: _myId),
                  ),
          ),

          _buildInput(),
        ],
      ),
    );
  }

  PreferredSizeWidget _buildAppBar() {
    return AppBar(
      backgroundColor: AppColors.primaryDark,
      leading: IconButton(
        icon: const Icon(Icons.arrow_back_ios_new,
            color: Colors.white, size: 18),
        onPressed: () => context.pop(),
      ),
      bottom: const PreferredSize(
        preferredSize: Size.fromHeight(1),
        child: Divider(height: 1, thickness: 1, color: AppColors.goldBorder),
      ),
      title: Column(
        children: [
          Text(
            'Махрам сөйлесу',
            style: GoogleFonts.nunito(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: Colors.white,
            ),
          ),
          Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              _dot(AppColors.bubbleMine),
              _dot(AppColors.bubbleMahramMale),
              _dot(AppColors.bubbleMahram),
              const SizedBox(width: 4),
              Text('3 қатысушы',
                  style: GoogleFonts.nunito(
                      fontSize: 10, color: AppColors.secondary)),
            ],
          ),
        ],
      ),
    );
  }

  Widget _dot(Color color) => Container(
        width: 8,
        height: 8,
        margin: const EdgeInsets.only(right: 4),
        decoration: BoxDecoration(color: color, shape: BoxShape.circle),
      );

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const IslamicStarWidget(size: 40, color: AppColors.secondary),
          const SizedBox(height: 16),
          Text(
            'Сөйлесуді бастаңыз',
            style: GoogleFonts.nunito(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary),
          ),
          const SizedBox(height: 8),
          Text(
            'Бұл сөйлесу барлық 3 тарапқа\n(әйел, ер, махрам) көрінеді',
            textAlign: TextAlign.center,
            style: GoogleFonts.nunito(
                fontSize: 13, color: AppColors.textSecondary),
          ),
        ],
      ),
    );
  }

  Widget _buildInput() {
    return SafeArea(
      child: Container(
        padding:
            const EdgeInsets.symmetric(horizontal: AppSpacing.md, vertical: 8),
        decoration: const BoxDecoration(
          color: AppColors.surface,
          boxShadow: [
            BoxShadow(
                color: Color(0x12000000),
                blurRadius: 8,
                offset: Offset(0, -2))
          ],
        ),
        child: Row(
          children: [
            Expanded(
              child: TextField(
                controller: _msgCtr,
                decoration: InputDecoration(
                  hintText: 'Хабарлама жазыңыз...',
                  hintStyle: GoogleFonts.nunito(
                      fontSize: 14, color: AppColors.textHint),
                  filled: true,
                  fillColor: AppColors.surfaceVariant,
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(20),
                    borderSide: BorderSide.none,
                  ),
                  contentPadding: const EdgeInsets.symmetric(
                      horizontal: 16, vertical: 10),
                ),
                style: GoogleFonts.nunito(
                    fontSize: 14, color: AppColors.textPrimary),
                textCapitalization: TextCapitalization.sentences,
                maxLines: 3,
                minLines: 1,
              ),
            ),
            const SizedBox(width: 8),
            GestureDetector(
              onTap: _send,
              child: Container(
                width: 44,
                height: 44,
                decoration: const BoxDecoration(
                  color: AppColors.primary,
                  shape: BoxShape.circle,
                ),
                child:
                    const Icon(Icons.send, color: Colors.white, size: 20),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ─── Mahram Bubble ────────────────────────────────────────────────────────────

class _MahramBubble extends StatelessWidget {
  final Map<String, dynamic> msg;
  final String myId;

  const _MahramBubble({required this.msg, required this.myId});

  @override
  Widget build(BuildContext context) {
    final senderId = msg['sender_id'] as String? ?? '';
    final role = msg['sender_role'] as String? ?? 'woman';
    final content = msg['content'] as String? ?? '';
    final isMe = senderId == myId;

    final bubbleColor = _bubbleColor(role, isMe);
    final textColor = isMe || role == 'man' || role == 'mahram'
        ? Colors.white
        : AppColors.textPrimary;
    final label = _roleLabel(role, isMe);

    return Padding(
      padding: const EdgeInsets.only(bottom: 10),
      child: Row(
        mainAxisAlignment:
            isMe ? MainAxisAlignment.end : MainAxisAlignment.start,
        crossAxisAlignment: CrossAxisAlignment.end,
        children: [
          if (!isMe) ...[
            CircleAvatar(
              radius: 14,
              backgroundColor: bubbleColor.withValues(alpha: 0.2),
              child: Text(
                label[0],
                style: GoogleFonts.nunito(
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                    color: bubbleColor),
              ),
            ),
            const SizedBox(width: 6),
          ],
          Flexible(
            child: Column(
              crossAxisAlignment:
                  isMe ? CrossAxisAlignment.end : CrossAxisAlignment.start,
              children: [
                if (!isMe)
                  Padding(
                    padding: const EdgeInsets.only(left: 4, bottom: 2),
                    child: Text(
                      label,
                      style: GoogleFonts.nunito(
                          fontSize: 11,
                          color: bubbleColor,
                          fontWeight: FontWeight.w600),
                    ),
                  ),
                Container(
                  constraints: BoxConstraints(
                    maxWidth: MediaQuery.of(context).size.width * 0.70,
                  ),
                  padding: const EdgeInsets.symmetric(
                      horizontal: 14, vertical: 10),
                  decoration: BoxDecoration(
                    color: bubbleColor,
                    borderRadius: BorderRadius.circular(18).copyWith(
                      bottomRight: isMe
                          ? const Radius.circular(4)
                          : const Radius.circular(18),
                      bottomLeft: !isMe
                          ? const Radius.circular(4)
                          : const Radius.circular(18),
                    ),
                  ),
                  child: Text(
                    content,
                    style: GoogleFonts.nunito(
                        fontSize: 15, color: textColor, height: 1.4),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Color _bubbleColor(String role, bool isMe) {
    if (isMe) return AppColors.bubbleMine;
    switch (role) {
      case 'man':
        return AppColors.bubbleMahramMale;
      case 'mahram':
        return AppColors.bubbleMahram;
      default:
        return AppColors.bubbleTheirs;
    }
  }

  String _roleLabel(String role, bool isMe) {
    if (isMe) return 'Сіз';
    switch (role) {
      case 'man':
        return 'Жігіт';
      case 'mahram':
        return 'Махрам';
      default:
        return 'Қыз';
    }
  }
}
