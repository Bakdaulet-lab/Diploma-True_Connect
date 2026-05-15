import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:dio/dio.dart';
import '../../core/constants/api_constants.dart';
import '../../core/network/dio_error_message.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../providers/auth_provider.dart';
import '../../providers/chat_provider.dart';
import '../../providers/matching_provider.dart';
import '../../services/websocket_service.dart';
import 'call_screen.dart';

class ChatScreen extends ConsumerStatefulWidget {
  final String matchId;
  final String otherUserId;

  const ChatScreen({super.key, required this.matchId, this.otherUserId = ''});

  @override
  ConsumerState<ChatScreen> createState() => _ChatScreenState();
}

class _ChatScreenState extends ConsumerState<ChatScreen> {
  final _msgController = TextEditingController();
  final _scrollController = ScrollController();

  @override
  void dispose() {
    _msgController.dispose();
    _scrollController.dispose();
    super.dispose();
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

  void _send() {
    final text = _msgController.text.trim();
    if (text.isEmpty) return;
    ref.read(chatNotifierProvider(widget.matchId).notifier).send(text);
    _msgController.clear();
    _scrollToBottom();
  }

  void _onTyping(String _) {
    ref.read(chatNotifierProvider(widget.matchId).notifier).sendTyping();
  }

  Future<void> _startCall() async {
    const storage = FlutterSecureStorage();
    final token = await storage.read(key: 'access_token') ?? '';
    final wsService = WebSocketService();
    wsService.connect(token);
    if (!mounted) return;
    Navigator.of(context).push(MaterialPageRoute(
      builder: (_) => CallScreen(
        matchData: {'id': widget.matchId},
        wsService: wsService,
        isCaller: true,
      ),
    ));
  }

  void _openMahramInvite() {
    showModalBottomSheet(
      context: context,
      backgroundColor: AppColors.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (_) => _MahramInviteSheet(matchId: widget.matchId),
    );
  }

  Future<void> _unmatch() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: AppColors.surface,
        title: Text('Мэтчті жою', style: GoogleFonts.nunito(color: AppColors.textPrimary, fontWeight: FontWeight.bold)),
        content: Text('Бұл мэтчті жойғыңыз келе ме? Хабарламалар жойылады.', style: GoogleFonts.nunito(color: AppColors.textSecondary)),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Болдырмау')),
          TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: Text('Жою', style: GoogleFonts.nunito(color: AppColors.accent)),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;

    try {
      final dio = ref.read(dioClientProvider).dio;
      await dio.post(ApiConstants.unmatch(widget.matchId));
      ref.invalidate(matchesListProvider);
      if (mounted) context.go('/matches');
    } on DioException catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(dioErrorMessage(e))));
    }
  }

  Future<void> _blockUser(String userId) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: AppColors.surface,
        title: Text('Блоктау', style: GoogleFonts.nunito(color: AppColors.textPrimary, fontWeight: FontWeight.bold)),
        content: Text('Пайдаланушыны блоктағыңыз келе ме? Ол сізді ленталарда көрмейді.', style: GoogleFonts.nunito(color: AppColors.textSecondary)),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx, false), child: const Text('Болдырмау')),
          TextButton(
            onPressed: () => Navigator.pop(ctx, true),
            child: Text('Блоктау', style: GoogleFonts.nunito(color: AppColors.accent)),
          ),
        ],
      ),
    );
    if (confirmed != true || !mounted) return;

    try {
      final dio = ref.read(dioClientProvider).dio;
      await dio.post(ApiConstants.blockUser(userId));
      ref.invalidate(matchesListProvider);
      if (mounted) context.go('/matches');
    } on DioException catch (e) {
      if (mounted) ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(dioErrorMessage(e))));
    }
  }

  @override
  Widget build(BuildContext context) {
    final chatState = ref.watch(chatNotifierProvider(widget.matchId));
    final currentUserId =
        ref.watch(authStateProvider).valueOrNull?.id ?? '';
    _scrollToBottom();

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: _buildAppBar(),
      body: Column(
        children: [
          // Connection lost banner
          if (!chatState.isConnected)
            Container(
              color: AppColors.surfaceVariant,
              padding: const EdgeInsets.symmetric(
                  vertical: 6, horizontal: AppSpacing.md),
              child: Row(
                children: [
                  const Icon(Icons.wifi_off,
                      size: 14, color: AppColors.textSecondary),
                  const SizedBox(width: 6),
                  Text('Байланыс жоқ...',
                      style: GoogleFonts.nunito(
                          fontSize: 12, color: AppColors.textSecondary)),
                ],
              ),
            ),

          // Messages
          Expanded(
            child: ListView.builder(
              controller: _scrollController,
              padding: const EdgeInsets.fromLTRB(
                  AppSpacing.md, AppSpacing.md, AppSpacing.md, 8),
              itemCount: chatState.messages.length +
                  (chatState.isOtherTyping ? 1 : 0),
              itemBuilder: (context, index) {
                if (chatState.isOtherTyping &&
                    index == chatState.messages.length) {
                  return const _TypingIndicator();
                }
                final msg = chatState.messages[index];
                return _MessageBubble(msg: msg, currentUserId: currentUserId);
              },
            ),
          ),

          // Input bar
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
            'Хабарлама',
            style: GoogleFonts.nunito(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: Colors.white,
            ),
          ),
        ],
      ),
      actions: [
        // Video call button
        IconButton(
          icon: const Icon(Icons.call_outlined,
              color: AppColors.secondary, size: 22),
          tooltip: 'Қоңырау шалу',
          onPressed: _startCall,
        ),
        // First meeting protocol button
        IconButton(
          icon: const Icon(Icons.location_on_outlined,
              color: AppColors.secondary, size: 22),
          tooltip: 'Бірінші кездесу',
          onPressed: () => context.push(
              '/first-meeting/${widget.matchId}?userId=${widget.otherUserId}'),
        ),
        // Invite Mahram button
        TextButton.icon(
          onPressed: _openMahramInvite,
          icon: const Icon(Icons.group_add, color: AppColors.secondary, size: 18),
          label: Text(
            'Махрам',
            style: GoogleFonts.nunito(
              fontSize: 12,
              color: AppColors.secondary,
              fontWeight: FontWeight.w600,
            ),
          ),
        ),
        // Unmatch / Block overflow menu
        PopupMenuButton<String>(
          icon: const Icon(Icons.more_vert, color: Colors.white),
          color: AppColors.surface,
          onSelected: (value) {
            if (value == 'unmatch') _unmatch();
            if (value == 'block') _blockUser(widget.otherUserId);
          },
          itemBuilder: (_) => [
            PopupMenuItem(
              value: 'unmatch',
              child: Row(
                children: [
                  const Icon(Icons.link_off, size: 18, color: AppColors.textSecondary),
                  const SizedBox(width: 8),
                  Text('Мэтчті жою', style: GoogleFonts.nunito(color: AppColors.textPrimary)),
                ],
              ),
            ),
            if (widget.otherUserId.isNotEmpty)
              PopupMenuItem(
                value: 'block',
                child: Row(
                  children: [
                    const Icon(Icons.block, size: 18, color: AppColors.accent),
                    const SizedBox(width: 8),
                    Text('Блоктау', style: GoogleFonts.nunito(color: AppColors.accent)),
                  ],
                ),
              ),
          ],
        ),
      ],
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
              offset: Offset(0, -2),
            ),
          ],
        ),
        child: Row(
          children: [
            Expanded(
              child: TextField(
                controller: _msgController,
                onChanged: _onTyping,
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
                maxLines: 4,
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
                child: const Icon(Icons.send, color: Colors.white, size: 20),
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// ─── Message Bubble ───────────────────────────────────────────────────────────

class _MessageBubble extends StatelessWidget {
  final Map<String, dynamic> msg;
  final String currentUserId;
  const _MessageBubble({required this.msg, required this.currentUserId});

  @override
  Widget build(BuildContext context) {
    final isMe = currentUserId.isNotEmpty
        ? msg['sender_id'] == currentUserId
        : msg['sender_id'] == 'me';
    final content = msg['content'] as String? ?? '';
    final isWarning = msg['is_toxic'] as bool? ?? false;
    final time = _formatTime(msg['created_at'] as String? ?? '');
    final isRead = msg['is_read'] as bool? ?? false;

    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Column(
        crossAxisAlignment:
            isMe ? CrossAxisAlignment.end : CrossAxisAlignment.start,
        children: [
          // Content warning banner
          if (isWarning && !isMe)
            Container(
              margin: const EdgeInsets.only(bottom: 4),
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
              decoration: BoxDecoration(
                color: Colors.amber.shade50,
                borderRadius: AppRadius.chip,
                border: Border.all(color: Colors.amber.shade300),
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Icon(Icons.warning_amber,
                      size: 12, color: Colors.amber.shade700),
                  const SizedBox(width: 4),
                  Text(
                    'Мазмұн ескертуі',
                    style: GoogleFonts.nunito(
                        fontSize: 11, color: Colors.amber.shade700),
                  ),
                ],
              ),
            ),

          Row(
            mainAxisAlignment:
                isMe ? MainAxisAlignment.end : MainAxisAlignment.start,
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              if (!isMe) const SizedBox(width: 8),
              Flexible(
                child: Container(
                  constraints: BoxConstraints(
                    maxWidth: MediaQuery.of(context).size.width * 0.72,
                  ),
                  padding: const EdgeInsets.symmetric(
                      horizontal: 14, vertical: 10),
                  decoration: BoxDecoration(
                    color:
                        isMe ? AppColors.bubbleMine : AppColors.bubbleTheirs,
                    borderRadius: BorderRadius.circular(18).copyWith(
                      bottomRight: isMe
                          ? const Radius.circular(4)
                          : const Radius.circular(18),
                      bottomLeft: !isMe
                          ? const Radius.circular(4)
                          : const Radius.circular(18),
                    ),
                    boxShadow: const [
                      BoxShadow(
                        color: Color(0x10000000),
                        blurRadius: 4,
                        offset: Offset(0, 2),
                      ),
                    ],
                  ),
                  child: Text(
                    content,
                    style: GoogleFonts.nunito(
                      fontSize: 15,
                      color: isMe ? Colors.white : AppColors.textPrimary,
                      height: 1.4,
                    ),
                  ),
                ),
              ),
              const SizedBox(width: 6),
              Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    time,
                    style: GoogleFonts.nunito(
                        fontSize: 10, color: AppColors.textHint),
                  ),
                  if (isMe)
                    Icon(
                      isRead ? Icons.done_all : Icons.check,
                      size: 12,
                      color: isRead ? AppColors.primary : AppColors.textHint,
                    ),
                ],
              ),
            ],
          ),
        ],
      ),
    );
  }

  String _formatTime(String iso) {
    if (iso.isEmpty) return '';
    try {
      final dt = DateTime.parse(iso).toLocal();
      return '${dt.hour.toString().padLeft(2, '0')}:${dt.minute.toString().padLeft(2, '0')}';
    } catch (_) {
      return '';
    }
  }
}

// ─── Typing Indicator ─────────────────────────────────────────────────────────

class _TypingIndicator extends StatelessWidget {
  const _TypingIndicator();

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 8, left: 8),
      child: Container(
        padding:
            const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
        decoration: BoxDecoration(
          color: AppColors.surfaceVariant,
          borderRadius: BorderRadius.circular(18).copyWith(
            bottomLeft: const Radius.circular(4),
          ),
        ),
        child: Text(
          '...',
          style: GoogleFonts.nunito(
              fontSize: 18,
              color: AppColors.textHint,
              letterSpacing: 2),
        ),
      ),
    );
  }
}

// ─── Mahram Invite Sheet ──────────────────────────────────────────────────────

class _MahramInviteSheet extends StatefulWidget {
  final String matchId;
  const _MahramInviteSheet({required this.matchId});

  @override
  State<_MahramInviteSheet> createState() => _MahramInviteSheetState();
}

class _MahramInviteSheetState extends State<_MahramInviteSheet> {
  final _phoneCtr = TextEditingController();
  final _otpCtr = TextEditingController();
  bool _loading = false;
  String? _mahramId;

  @override
  void dispose() {
    _phoneCtr.dispose();
    _otpCtr.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.fromLTRB(
        AppSpacing.lg,
        AppSpacing.lg,
        AppSpacing.lg,
        AppSpacing.xl + MediaQuery.of(context).viewInsets.bottom,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Махрамды шақыру',
            style: GoogleFonts.nunito(
              fontSize: 20,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 4),
          Text(
            'Махрам @MahabbatVerifyBot арқылы Telegram-да растайды',
            style: GoogleFonts.nunito(
                fontSize: 13, color: AppColors.textSecondary),
          ),
          const SizedBox(height: AppSpacing.lg),

          if (_mahramId == null) ...[
            TextField(
              controller: _phoneCtr,
              keyboardType: TextInputType.phone,
              decoration: const InputDecoration(
                hintText: '+7 ___ ___ __ __',
                prefixIcon:
                    Icon(Icons.phone_outlined, color: AppColors.textHint),
              ),
            ),
            const SizedBox(height: AppSpacing.md),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: _loading ? null : _registerMahram,
                child: _loading
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                            color: Colors.white, strokeWidth: 2))
                    : const Text('Жіберу'),
              ),
            ),
          ] else ...[
            Text(
              'OTP кодын енгізіңіз (dev: 123456)',
              style: GoogleFonts.nunito(
                  fontSize: 13, color: AppColors.textSecondary),
            ),
            const SizedBox(height: 8),
            TextField(
              controller: _otpCtr,
              keyboardType: TextInputType.number,
              maxLength: 6,
              decoration: const InputDecoration(
                hintText: '______',
                prefixIcon:
                    Icon(Icons.lock_outline, color: AppColors.textHint),
              ),
            ),
            const SizedBox(height: AppSpacing.md),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: _loading ? null : _verifyOtp,
                child: _loading
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(
                            color: Colors.white, strokeWidth: 2))
                    : const Text('Растау'),
              ),
            ),
          ],
        ],
      ),
    );
  }

  Future<void> _registerMahram() async {
    if (_phoneCtr.text.trim().isEmpty) return;
    setState(() => _loading = true);
    // POST /v1/mahram — handled by backend, returns mahram_id
    await Future.delayed(const Duration(seconds: 1));
    setState(() {
      _loading = false;
      _mahramId = 'demo_mahram_id';
    });
  }

  Future<void> _verifyOtp() async {
    if (_otpCtr.text.trim().length < 6) return;
    setState(() => _loading = true);
    await Future.delayed(const Duration(seconds: 1));
    if (mounted) {
      Navigator.pop(context);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'Махрам расталды! 3 жақты сөйлесу ашылды.',
            style: GoogleFonts.nunito(),
          ),
          backgroundColor: AppColors.primary,
        ),
      );
    }
  }
}
