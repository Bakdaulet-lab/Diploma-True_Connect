import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:google_fonts/google_fonts.dart';
import '../core/theme/app_colors.dart';
import '../core/theme/app_theme.dart';
import '../models/post.dart';

/// Threads-style post card: avatar in the left gutter, the author name +
/// inline timestamp + overflow menu on the header row, and the post body and
/// actions aligned under the name (not the avatar).
class PostCard extends StatelessWidget {
  final Post post;
  final VoidCallback? onLike;
  final VoidCallback? onComment;
  final VoidCallback? onShare;
  final VoidCallback? onTap;

  const PostCard({
    super.key,
    required this.post,
    this.onLike,
    this.onComment,
    this.onShare,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      behavior: HitTestBehavior.opaque,
      child: Container(
        color: AppColors.surface,
        padding: const EdgeInsets.fromLTRB(
            AppSpacing.md, AppSpacing.md, AppSpacing.md, AppSpacing.sm),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            _Avatar(avatarUrl: post.authorAvatarUrl, name: post.authorName),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _HeaderRow(post: post),
                  if (post.content.isNotEmpty) ...[
                    const SizedBox(height: 3),
                    Text(
                      post.content,
                      style: GoogleFonts.nunito(
                        fontSize: 15,
                        color: AppColors.textPrimary,
                        height: 1.45,
                      ),
                    ),
                  ],
                  if (post.mediaUrl != null && post.mediaUrl!.isNotEmpty) ...[
                    const SizedBox(height: 10),
                    ClipRRect(
                      borderRadius: BorderRadius.circular(14),
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
                  ],
                  const SizedBox(height: 6),
                  _ActionRow(
                      post: post,
                      onLike: onLike,
                      onComment: onComment,
                      onShare: onShare),
                  if (post.likeCount > 0 || post.commentCount > 0)
                    Padding(
                      padding: const EdgeInsets.only(top: 2),
                      child: Text(
                        _summary(post),
                        style: GoogleFonts.nunito(
                          fontSize: 12.5,
                          color: AppColors.textHint,
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  String _summary(Post post) {
    final parts = <String>[];
    if (post.commentCount > 0) parts.add('${post.commentCount} пікір');
    if (post.likeCount > 0) parts.add('${post.likeCount} ұнату');
    return parts.join(' · ');
  }
}

class _HeaderRow extends StatelessWidget {
  final Post post;
  const _HeaderRow({required this.post});

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        Flexible(
          child: Text(
            post.authorName,
            style: GoogleFonts.nunito(
              fontSize: 15,
              fontWeight: FontWeight.w700,
              color: AppColors.textPrimary,
            ),
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),
        ),
        const SizedBox(width: 6),
        Text(
          _formatTime(post.createdAt),
          style: GoogleFonts.nunito(
            fontSize: 13,
            color: AppColors.textHint,
          ),
        ),
        const Spacer(),
        _OverflowMenu(post: post),
      ],
    );
  }

  String _formatTime(DateTime dt) {
    final diff = DateTime.now().difference(dt);
    if (diff.inMinutes < 1) return 'Жаңа ғана';
    if (diff.inHours < 1) return '${diff.inMinutes} мин';
    if (diff.inDays < 1) return '${diff.inHours} сағ';
    if (diff.inDays < 7) return '${diff.inDays} күн';
    return '${dt.day}.${dt.month.toString().padLeft(2, '0')}.${dt.year}';
  }
}

class _OverflowMenu extends StatelessWidget {
  final Post post;
  const _OverflowMenu({required this.post});

  @override
  Widget build(BuildContext context) {
    return PopupMenuButton<String>(
      icon: const Icon(Icons.more_horiz, size: 20, color: AppColors.textHint),
      padding: EdgeInsets.zero,
      splashRadius: 18,
      color: AppColors.surface,
      shape: const RoundedRectangleBorder(borderRadius: AppRadius.input),
      onSelected: (value) {
        if (value == 'copy') {
          Clipboard.setData(ClipboardData(text: post.content));
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Мәтін көшірілді'),
              backgroundColor: AppColors.primary,
              behavior: SnackBarBehavior.floating,
            ),
          );
        }
      },
      itemBuilder: (_) => [
        PopupMenuItem(
          value: 'copy',
          child: Row(
            children: [
              const Icon(Icons.copy_outlined,
                  size: 18, color: AppColors.textSecondary),
              const SizedBox(width: 10),
              Text('Мәтінді көшіру',
                  style: GoogleFonts.nunito(
                      fontSize: 14, color: AppColors.textPrimary)),
            ],
          ),
        ),
      ],
    );
  }
}

class _Avatar extends StatelessWidget {
  final String? avatarUrl;
  final String name;
  const _Avatar({this.avatarUrl, required this.name});

  @override
  Widget build(BuildContext context) {
    return CircleAvatar(
      radius: 22,
      backgroundColor: AppColors.primaryLight,
      backgroundImage: avatarUrl != null && avatarUrl!.isNotEmpty
          ? CachedNetworkImageProvider(avatarUrl!)
          : null,
      child: avatarUrl == null || avatarUrl!.isEmpty
          ? Text(
              name.isNotEmpty ? name[0].toUpperCase() : '?',
              style: GoogleFonts.nunito(
                fontSize: 17,
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
  final VoidCallback? onShare;

  const _ActionRow(
      {required this.post, this.onLike, this.onComment, this.onShare});

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        _AnimatedLikeButton(
          isLiked: post.isLiked,
          count: post.likeCount,
          onTap: onLike,
        ),
        const SizedBox(width: AppSpacing.sm),
        _CommentButton(count: post.commentCount, onTap: onComment),
        const SizedBox(width: AppSpacing.sm),
        _ShareButton(onTap: onShare),
      ],
    );
  }
}

class _ShareButton extends StatelessWidget {
  final VoidCallback? onTap;
  const _ShareButton({this.onTap});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      behavior: HitTestBehavior.opaque,
      child: const Padding(
        padding: EdgeInsets.symmetric(
            horizontal: AppSpacing.sm, vertical: AppSpacing.sm),
        child: Icon(Icons.send_outlined,
            size: 20, color: AppColors.textSecondary),
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
      behavior: HitTestBehavior.opaque,
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
                size: 23,
                color: widget.isLiked ? Colors.red : AppColors.textSecondary,
              ),
            ),
            if (widget.count > 0) ...[
              const SizedBox(width: 6),
              Text(
                '${widget.count}',
                style: GoogleFonts.nunito(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color:
                      widget.isLiked ? Colors.red : AppColors.textSecondary,
                ),
              ),
            ],
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
      behavior: HitTestBehavior.opaque,
      child: Padding(
        padding: const EdgeInsets.symmetric(
            horizontal: AppSpacing.sm, vertical: AppSpacing.sm),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(Icons.mode_comment_outlined,
                size: 21, color: AppColors.textSecondary),
            if (count > 0) ...[
              const SizedBox(width: 6),
              Text(
                '$count',
                style: GoogleFonts.nunito(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: AppColors.textSecondary,
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }
}
