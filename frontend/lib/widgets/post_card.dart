import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../core/theme/app_colors.dart';
import '../core/theme/app_theme.dart';
import '../models/post.dart';

class PostCard extends StatelessWidget {
  final Post post;
  final VoidCallback? onLike;
  final VoidCallback? onComment;
  final VoidCallback? onTap;

  const PostCard({
    super.key,
    required this.post,
    this.onLike,
    this.onComment,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        margin: const EdgeInsets.symmetric(
            horizontal: AppSpacing.md, vertical: AppSpacing.xs),
        decoration: const BoxDecoration(
          color: AppColors.surface,
          borderRadius: BorderRadius.all(Radius.circular(16)),
          boxShadow: AppShadows.soft,
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _AuthorRow(post: post),
            Padding(
              padding: const EdgeInsets.symmetric(
                  horizontal: AppSpacing.md, vertical: AppSpacing.sm),
              child: Text(
                post.content,
                style: GoogleFonts.nunito(
                  fontSize: 14,
                  color: AppColors.textPrimary,
                  height: 1.55,
                ),
              ),
            ),
            if (post.mediaUrl != null && post.mediaUrl!.isNotEmpty)
              ClipRRect(
                borderRadius: const BorderRadius.only(
                  bottomLeft: Radius.circular(16),
                  bottomRight: Radius.circular(16),
                ),
                child: Image.network(
                  post.mediaUrl!,
                  width: double.infinity,
                  height: 200,
                  fit: BoxFit.cover,
                  errorBuilder: (_, __, ___) => const SizedBox.shrink(),
                ),
              ),
            _ActionRow(post: post, onLike: onLike, onComment: onComment),
          ],
        ),
      ),
    );
  }
}

class _AuthorRow extends StatelessWidget {
  final Post post;
  const _AuthorRow({required this.post});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(
          AppSpacing.md, AppSpacing.md, AppSpacing.md, 0),
      child: Row(
        children: [
          CircleAvatar(
            radius: 18,
            backgroundColor: AppColors.primaryLight,
            backgroundImage: post.authorAvatarUrl != null
                ? NetworkImage(post.authorAvatarUrl!)
                : null,
            child: post.authorAvatarUrl == null
                ? Text(
                    post.authorName.isNotEmpty
                        ? post.authorName[0].toUpperCase()
                        : '?',
                    style: GoogleFonts.nunito(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color: AppColors.primary,
                    ),
                  )
                : null,
          ),
          const SizedBox(width: AppSpacing.sm),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  post.authorName,
                  style: GoogleFonts.nunito(
                    fontSize: 13,
                    fontWeight: FontWeight.w700,
                    color: AppColors.textPrimary,
                  ),
                ),
                Text(
                  _formatTime(post.createdAt),
                  style: GoogleFonts.nunito(
                    fontSize: 11,
                    color: AppColors.textHint,
                  ),
                ),
              ],
            ),
          ),
        ],
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

class _ActionRow extends StatelessWidget {
  final Post post;
  final VoidCallback? onLike;
  final VoidCallback? onComment;

  const _ActionRow({required this.post, this.onLike, this.onComment});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(
          horizontal: AppSpacing.sm, vertical: AppSpacing.xs),
      child: Row(
        children: [
          _ActionButton(
            icon: post.isLiked ? Icons.favorite : Icons.favorite_border,
            color: post.isLiked ? AppColors.accent : AppColors.textHint,
            count: post.likeCount,
            onTap: onLike,
          ),
          const SizedBox(width: AppSpacing.sm),
          _ActionButton(
            icon: Icons.chat_bubble_outline,
            color: AppColors.textHint,
            count: post.commentCount,
            onTap: onComment,
          ),
        ],
      ),
    );
  }
}

class _ActionButton extends StatelessWidget {
  final IconData icon;
  final Color color;
  final int count;
  final VoidCallback? onTap;

  const _ActionButton(
      {required this.icon,
      required this.color,
      required this.count,
      this.onTap});

  @override
  Widget build(BuildContext context) {
    return InkWell(
      borderRadius: const BorderRadius.all(Radius.circular(8)),
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.symmetric(
            horizontal: AppSpacing.sm, vertical: AppSpacing.xs),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 18, color: color),
            const SizedBox(width: 4),
            Text(
              '$count',
              style: GoogleFonts.nunito(
                  fontSize: 12,
                  color: color,
                  fontWeight: FontWeight.w600),
            ),
          ],
        ),
      ),
    );
  }
}
