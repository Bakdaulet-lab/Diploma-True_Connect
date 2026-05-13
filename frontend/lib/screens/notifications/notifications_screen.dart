import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../models/notification.dart';
import '../../providers/notification_provider.dart';
import '../../widgets/common_widgets.dart';
import '../../widgets/empty_state.dart';

class NotificationsScreen extends ConsumerWidget {
  const NotificationsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(notificationsProvider);

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        title: Text(
          'Хабарландырулар',
          style: GoogleFonts.nunito(
            fontSize: 18,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
        ),
        actions: [
          TextButton(
            onPressed: () =>
                ref.read(notificationsProvider.notifier).markAllRead(),
            child: Text(
              'Барлығын оқу',
              style: GoogleFonts.nunito(
                fontSize: 13,
                color: Colors.white70,
              ),
            ),
          ),
        ],
      ),
      body: state.when(
        loading: () => const LoadingWidget(message: 'Жүктелуде...'),
        error: (e, _) => ErrorRetryWidget(
          message: e.toString(),
          onRetry: () => ref.read(notificationsProvider.notifier).load(),
        ),
        data: (notifications) {
          if (notifications.isEmpty) {
            return const EmptyState(
              icon: Icons.notifications_off_outlined,
              title: 'Хабарландыру жоқ',
              subtitle: 'Жаңа хабарландырулар осында көрінеді',
            );
          }
          return RefreshIndicator(
            color: AppColors.primary,
            onRefresh: () =>
                ref.read(notificationsProvider.notifier).load(),
            child: ListView.separated(
              padding: const EdgeInsets.symmetric(vertical: AppSpacing.sm),
              itemCount: notifications.length,
              separatorBuilder: (_, __) =>
                  const Divider(height: 1, indent: 72),
              itemBuilder: (context, i) => _NotificationTile(
                notification: notifications[i],
                onTap: () {
                  ref
                      .read(notificationsProvider.notifier)
                      .markRead(notifications[i].id);
                  _navigate(context, notifications[i]);
                },
              ),
            ),
          );
        },
      ),
    );
  }

  void _navigate(BuildContext context, AppNotification n) {
    final type = n.type;
    final matchId = n.data['match_id'] as String?;
    final roomId = n.data['room_id'] as String?;

    switch (type) {
      case 'new_match':
        context.push('/matches');
      case 'new_message':
        if (matchId != null) context.push('/chat/$matchId');
      case 'mahram_message':
        if (roomId != null) context.push('/mahram-chat/$roomId');
      case 'post_like':
      case 'post_comment':
        context.push('/feed');
    }
  }
}

class _NotificationTile extends StatelessWidget {
  final AppNotification notification;
  final VoidCallback onTap;

  const _NotificationTile({
    required this.notification,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Container(
        color: notification.isRead
            ? Colors.transparent
            : AppColors.primaryLight.withValues(alpha: 0.5),
        padding: const EdgeInsets.symmetric(
            horizontal: AppSpacing.md, vertical: AppSpacing.sm),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _TypeIcon(type: notification.type),
            const SizedBox(width: AppSpacing.md),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (notification.title.isNotEmpty)
                    Text(
                      notification.title,
                      style: GoogleFonts.nunito(
                        fontSize: 13,
                        fontWeight: notification.isRead
                            ? FontWeight.normal
                            : FontWeight.w700,
                        color: AppColors.textPrimary,
                      ),
                    ),
                  if (notification.body.isNotEmpty)
                    Text(
                      notification.body,
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                      style: GoogleFonts.nunito(
                        fontSize: 12,
                        color: AppColors.textSecondary,
                        height: 1.4,
                      ),
                    ),
                  const SizedBox(height: 2),
                  Text(
                    _formatTime(notification.createdAt),
                    style: GoogleFonts.nunito(
                      fontSize: 11,
                      color: AppColors.textHint,
                    ),
                  ),
                ],
              ),
            ),
            if (!notification.isRead)
              Container(
                width: 8,
                height: 8,
                margin: const EdgeInsets.only(top: 4),
                decoration: const BoxDecoration(
                  color: AppColors.primary,
                  shape: BoxShape.circle,
                ),
              ),
          ],
        ),
      ),
    );
  }

  String _formatTime(DateTime dt) {
    final diff = DateTime.now().difference(dt);
    if (diff.inMinutes < 1) return 'Жаңа ғана';
    if (diff.inHours < 1) return '${diff.inMinutes} мин бұрын';
    if (diff.inDays < 1) return '${diff.inHours} сағ бұрын';
    if (diff.inDays < 7) return '${diff.inDays} күн бұрын';
    return '${dt.day}.${dt.month.toString().padLeft(2, '0')}.${dt.year}';
  }
}

class _TypeIcon extends StatelessWidget {
  final String type;
  const _TypeIcon({required this.type});

  @override
  Widget build(BuildContext context) {
    final (icon, color) = switch (type) {
      'new_match' => (Icons.favorite, AppColors.accent),
      'new_message' => (Icons.chat_bubble_outline, AppColors.primary),
      'mahram_message' => (Icons.people_outline, AppColors.secondary),
      'post_like' => (Icons.thumb_up_outlined, AppColors.primary),
      'post_comment' => (Icons.comment_outlined, AppColors.primary),
      _ => (Icons.notifications_outlined, AppColors.textHint),
    };

    return Container(
      width: 40,
      height: 40,
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        shape: BoxShape.circle,
      ),
      child: Icon(icon, size: 20, color: color),
    );
  }
}
