import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../core/theme/app_colors.dart';
import '../core/theme/app_theme.dart';

class MessageBubble extends StatelessWidget {
  final Map<String, dynamic> msg;

  /// Override bubble color for named roles (e.g. mahram chat: 'mahram' → gold).
  final Color? overrideBubbleColor;
  final Color? overrideTextColor;

  const MessageBubble({
    super.key,
    required this.msg,
    this.overrideBubbleColor,
    this.overrideTextColor,
  });

  @override
  Widget build(BuildContext context) {
    final isMe = msg['sender_id'] == 'me';
    final content = msg['content'] as String? ?? '';
    final isWarning = msg['is_toxic'] as bool? ?? false;
    final time = _formatTime(msg['created_at'] as String? ?? '');
    final isRead = msg['is_read'] as bool? ?? false;

    final bubbleColor = overrideBubbleColor ??
        (isMe ? AppColors.bubbleMine : AppColors.bubbleTheirs);
    final textColor = overrideTextColor ??
        (isMe ? Colors.white : AppColors.textPrimary);

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
              padding:
                  const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
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
                    color: bubbleColor,
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
                      color: textColor,
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
                      color: isRead
                          ? AppColors.primary
                          : AppColors.textHint,
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
      return '${dt.hour.toString().padLeft(2, '0')}:'
          '${dt.minute.toString().padLeft(2, '0')}';
    } catch (_) {
      return '';
    }
  }
}
