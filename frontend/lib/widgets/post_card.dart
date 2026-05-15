import 'package:cached_network_image/cached_network_image.dart';
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
        color: AppColors.surface,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _AuthorRow(post: post),
            if (post.content.isNotEmpty)
              Padding(
                padding: const EdgeInsets.fromLTRB(
                    AppSpacing.md, AppSpacing.xs, AppSpacing.md, AppSpacing.sm),
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
              Padding(
                padding: const EdgeInsets.symmetric(
                    horizontal: AppSpacing.md, vertical: AppSpacing.xs),
                child: ClipRRect(
                  borderRadius: const BorderRadius.all(Radius.circular(12)),
                  child: CachedNetworkImage(
                    imageUrl: post.mediaUrl!,
                    width: double.infinity,
                    fit: BoxFit.cover,
                    placeholder: (_, __) => Container(
                      height: 200,
                      color: AppColors.surfaceVariant,
                    ),
                    errorWidget: (_, __, ___) => const SizedBox.shrink(),
                  ),
                ),
              ),
            _ActionRow(post: post, onLike: onLike, onComment: onComment),
            const Divider(height: 1, thickness: 0.5, color: AppColors.divider),
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
          AppSpacing.md, AppSpacing.md, AppSpacing.md, AppSpacing.xs),
      child: Row(
        children: [
          _Avatar(
            avatarUrl: post.authorAvatarUrl,
            name: post.authorName,
          ),
          const SizedBox(width: 10),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  post.authorName,
                  style: GoogleFonts.nunito(
                    fontSize: 14,
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

class _Avatar extends StatelessWidget {
  final String? avatarUrl;
  final String name;
  const _Avatar({this.avatarUrl, required this.name});

  @override
  Widget build(BuildContext context) {
    return CircleAvatar(
      radius: 20,
      backgroundColor: AppColors.primaryLight,
      backgroundImage:
          avatarUrl != null && avatarUrl!.isNotEmpty
              ? CachedNetworkImageProvider(avatarUrl!)
              : null,
      child: avatarUrl == null || avatarUrl!.isEmpty
          ? Text(
              name.isNotEmpty ? name[0].toUpperCase() : '?',
              style: GoogleFonts.nunito(
                fontSize: 15,
                fontWeight: FontWeight.bold,
                color: AppColors.primary,
              ),
            )
          : null,
    );
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
          _AnimatedLikeButton(
            isLiked: post.isLiked,
            count: post.likeCount,
            onTap: onLike,
          ),
          const SizedBox(width: AppSpacing.md),
          _CommentButton(count: post.commentCount, onTap: onComment),
        ],
      ),
    );
  }
}

class _AnimatedLikeButton extends StatefulWidget {
  final bool isLiked;
  final int count;
  final VoidCallback? onTap;

  const _AnimatedLikeButton({
    required this.isLiked,
    required this.count,
    this.onTap,
  });

  @override
  State<_AnimatedLikeButton> createState() => _AnimatedLikeButtonState();
}

class _AnimatedLikeButtonState extends State<_AnimatedLikeButton>
    with SingleTickerProviderStateMixin {
  late final AnimationController _ctrl;
  late final Animation<double> _scale;

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(
        vsync: this, duration: const Duration(milliseconds: 350));
    _scale = TweenSequence([
      TweenSequenceItem(tween: Tween(begin: 1.0, end: 1.4), weight: 40),
      TweenSequenceItem(
          tween: Tween(begin: 1.4, end: 0.9)
              .chain(CurveTween(curve: Curves.easeOut)),
          weight: 30),
      TweenSequenceItem(
          tween: Tween(begin: 0.9, end: 1.0)
              .chain(CurveTween(curve: Curves.easeIn)),
          weight: 30),
    ]).animate(_ctrl);
  }

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  @override
  void didUpdateWidget(_AnimatedLikeButton old) {
    super.didUpdateWidget(old);
    if (!old.isLiked && widget.isLiked) {
      _ctrl.forward(from: 0);
    }
  }

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: () {
        if (!widget.isLiked) _ctrl.forward(from: 0);
        widget.onTap?.call();
      },
      child: Padding(
        padding: const EdgeInsets.symmetric(
            horizontal: AppSpacing.sm, vertical: AppSpacing.sm),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            ScaleTransition(
              scale: _scale,
              child: Icon(
                widget.isLiked ? Icons.favorite : Icons.favorite_border,
                size: 22,
                color: widget.isLiked ? Colors.red : AppColors.textHint,
              ),
            ),
            const SizedBox(width: 5),
            Text(
              '${widget.count}',
              style: GoogleFonts.nunito(
                fontSize: 13,
                fontWeight: FontWeight.w600,
                color: widget.isLiked ? Colors.red : AppColors.textHint,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _CommentButton extends StatelessWidget {
  final int count;
  final VoidCallback? onTap;
  const _CommentButton({required this.count, this.onTap});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Padding(
        padding: const EdgeInsets.symmetric(
            horizontal: AppSpacing.sm, vertical: AppSpacing.sm),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.chat_bubble_outline,
                size: 20, color: AppColors.textHint),
            const SizedBox(width: 5),
            Text(
              '$count',
              style: GoogleFonts.nunito(
                fontSize: 13,
                fontWeight: FontWeight.w600,
                color: AppColors.textHint,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
