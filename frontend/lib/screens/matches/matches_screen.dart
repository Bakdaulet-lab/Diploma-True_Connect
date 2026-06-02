import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../providers/matching_provider.dart';
import '../../widgets/halal_pattern_painter.dart';
import '../../widgets/niyyah_badge.dart';
import '../../widgets/trust_score_badge.dart';

class MatchesScreen extends ConsumerWidget {
  const MatchesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncMatches = ref.watch(matchesListProvider);
    final asyncLikes = ref.watch(pendingLikesProvider);

    final likeCount =
        asyncLikes.valueOrNull?.length ?? 0;

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.primaryDark,
        centerTitle: true,
        bottom: const PreferredSize(
          preferredSize: Size.fromHeight(1),
          child: Divider(height: 1, thickness: 1, color: AppColors.goldBorder),
        ),
        title: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              'Сәйкестіктер',
              style: GoogleFonts.nunito(
                fontSize: 20,
                fontWeight: FontWeight.bold,
                color: Colors.white,
              ),
            ),
            if (likeCount > 0) ...[
              const SizedBox(width: 8),
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 7, vertical: 2),
                decoration: BoxDecoration(
                  color: AppColors.accent,
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Text(
                  '$likeCount',
                  style: GoogleFonts.nunito(
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
              ),
            ],
          ],
        ),
      ),
      body: RefreshIndicator(
        color: AppColors.primary,
        onRefresh: () => Future.wait([
          ref.refresh(matchesListProvider.future),
          ref.refresh(pendingLikesProvider.future),
        ]),
        child: asyncMatches.when(
          loading: () => const Center(
              child: CircularProgressIndicator(color: AppColors.primary)),
          error: (e, _) => _buildError(context, ref, e.toString()),
          data: (matches) => CustomScrollView(
            slivers: [
              // ── "Liked You" section ──────────────────────────────────────
              SliverToBoxAdapter(
                child: _LikedYouSection(asyncLikes: asyncLikes),
              ),

              // ── Matches list ─────────────────────────────────────────────
              if (matches.isEmpty)
                SliverFillRemaining(child: _buildEmpty())
              else
                SliverPadding(
                  padding: const EdgeInsets.fromLTRB(
                      AppSpacing.md, 0, AppSpacing.md, AppSpacing.xxl),
                  sliver: SliverList.separated(
                    itemCount: matches.length,
                    separatorBuilder: (_, __) =>
                        const SizedBox(height: AppSpacing.sm),
                    itemBuilder: (context, index) =>
                        _MatchCard(match: matches[index]),
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildEmpty() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const IslamicStarWidget(size: 56, color: AppColors.secondary),
          const SizedBox(height: 24),
          Text(
            'Сәйкестіктер жоқ',
            style: GoogleFonts.nunito(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Жаңа адамдарды танып, Ұнайды батырмасын басыңыз',
            style:
                GoogleFonts.nunito(fontSize: 14, color: AppColors.textSecondary),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }

  Widget _buildError(BuildContext context, WidgetRef ref, String msg) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.wifi_off, size: 48, color: AppColors.textHint),
          const SizedBox(height: 16),
          Text(msg,
              style: GoogleFonts.nunito(color: AppColors.textSecondary),
              textAlign: TextAlign.center),
          const SizedBox(height: 16),
          ElevatedButton(
            onPressed: () => ref.refresh(matchesListProvider.future),
            child: const Text('Қайтадан көру'),
          ),
        ],
      ),
    );
  }
}

// ─── Liked You Section ────────────────────────────────────────────────────────

class _LikedYouSection extends ConsumerWidget {
  final AsyncValue<List<Map<String, dynamic>>> asyncLikes;
  const _LikedYouSection({required this.asyncLikes});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final likers = asyncLikes.valueOrNull;
    if (likers == null || likers.isEmpty) return const SizedBox.shrink();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.fromLTRB(
              AppSpacing.md, AppSpacing.md, AppSpacing.md, AppSpacing.sm),
          child: Row(
            children: [
              const Icon(Icons.favorite,
                  size: 16, color: AppColors.accent),
              const SizedBox(width: 6),
              Text(
                'Сізді ұнатқандар',
                style: GoogleFonts.nunito(
                  fontSize: 15,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
              const SizedBox(width: 6),
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 7, vertical: 1),
                decoration: BoxDecoration(
                  color: AppColors.accent,
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Text(
                  '${likers.length}',
                  style: GoogleFonts.nunito(
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
              ),
            ],
          ),
        ),
        SizedBox(
          height: 170,
          child: ListView.builder(
            scrollDirection: Axis.horizontal,
            padding: const EdgeInsets.symmetric(horizontal: AppSpacing.md),
            itemCount: likers.length,
            itemBuilder: (context, i) =>
                _LikerCard(liker: likers[i]),
          ),
        ),
        const Padding(
          padding: EdgeInsets.symmetric(vertical: AppSpacing.sm),
          child: Divider(color: AppColors.divider, height: 1),
        ),
        if (asyncLikes.valueOrNull?.isNotEmpty == true)
          Padding(
            padding: const EdgeInsets.fromLTRB(
                AppSpacing.md, 0, AppSpacing.md, AppSpacing.sm),
            child: Text(
              'Сәйкестіктер',
              style: GoogleFonts.nunito(
                fontSize: 15,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
          ),
      ],
    );
  }
}

class _LikerCard extends ConsumerWidget {
  final Map<String, dynamic> liker;
  const _LikerCard({required this.liker});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final userId = liker['id'] as String? ?? '';
    final name = liker['display_name'] as String? ??
        liker['other_user_name'] as String? ?? '';
    final avatarUrl = liker['avatar_url'] as String? ??
        liker['other_user_avatar_url'] as String?;
    final trustScore =
        (liker['trust_score'] as num?)?.toInt() ??
        (liker['other_user_trust_score'] as num?)?.toInt() ?? 0;

    return Container(
      width: 120,
      margin: const EdgeInsets.only(right: AppSpacing.sm),
      decoration: const BoxDecoration(
        color: AppColors.surface,
        borderRadius: AppRadius.card,
        boxShadow: AppShadows.soft,
      ),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Stack(
            alignment: Alignment.bottomRight,
            children: [
              CircleAvatar(
                radius: 34,
                backgroundColor: AppColors.primaryLight,
                backgroundImage: avatarUrl != null && avatarUrl.isNotEmpty
                    ? CachedNetworkImageProvider(avatarUrl)
                    : null,
                child: avatarUrl == null || avatarUrl.isEmpty
                    ? Text(
                        name.isNotEmpty ? name[0].toUpperCase() : '?',
                        style: GoogleFonts.nunito(
                          fontSize: 22,
                          fontWeight: FontWeight.bold,
                          color: AppColors.primary,
                        ),
                      )
                    : null,
              ),
              TrustScoreBadge(score: trustScore, size: 20),
            ],
          ),
          const SizedBox(height: 6),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 6),
            child: Text(
              name,
              style: GoogleFonts.nunito(
                fontSize: 12,
                fontWeight: FontWeight.w600,
                color: AppColors.textPrimary,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              textAlign: TextAlign.center,
            ),
          ),
          const SizedBox(height: 6),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              // Pass
              _IconAction(
                icon: Icons.close,
                color: AppColors.accent,
                onTap: () async {
                  await ref
                      .read(matchingNotifierProvider.notifier)
                      .pass(userId);
                  ref.invalidate(pendingLikesProvider);
                },
              ),
              const SizedBox(width: 8),
              // Like back
              _IconAction(
                icon: Icons.favorite,
                color: AppColors.primary,
                onTap: () async {
                  final matched = await ref
                      .read(matchingNotifierProvider.notifier)
                      .like(userId);
                  ref.invalidate(pendingLikesProvider);
                  if (matched && context.mounted) {
                    ref.invalidate(matchesListProvider);
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(
                        content: Text(
                          '🎉 $name — жаңа сәйкестік!',
                          style: GoogleFonts.nunito(
                              fontWeight: FontWeight.w600),
                        ),
                        backgroundColor: AppColors.primary,
                        behavior: SnackBarBehavior.floating,
                      ),
                    );
                  }
                },
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _IconAction extends StatelessWidget {
  final IconData icon;
  final Color color;
  final VoidCallback onTap;
  const _IconAction(
      {required this.icon, required this.color, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: 32,
        height: 32,
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.1),
          shape: BoxShape.circle,
          border: Border.all(color: color.withValues(alpha: 0.3)),
        ),
        child: Icon(icon, size: 16, color: color),
      ),
    );
  }
}

// ─── Match Card ───────────────────────────────────────────────────────────────

class _MatchCard extends ConsumerStatefulWidget {
  final Map<String, dynamic> match;
  const _MatchCard({required this.match});

  @override
  ConsumerState<_MatchCard> createState() => _MatchCardState();
}

class _MatchCardState extends ConsumerState<_MatchCard> {
  bool _introLoading = false;

  Future<void> _doFamilyIntro(String matchId) async {
    setState(() => _introLoading = true);
    try {
      await markFamilyIntro(ref, matchId);
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Отбасы таныстырылды! +15 Trust Score'),
          backgroundColor: AppColors.primary,
        ),
      );
      ref.invalidate(matchesListProvider);
    } catch (_) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Қате орын алды, қайталап көріңіз')),
      );
    } finally {
      if (mounted) setState(() => _introLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final name = widget.match['other_user_name'] as String? ?? '';
    final avatarUrl = widget.match['other_user_avatar_url'] as String?;
    final trustScore =
        (widget.match['other_user_trust_score'] as num?)?.toInt() ?? 0;
    final matchId = widget.match['id'] as String? ?? '';
    final otherUserId = widget.match['other_user_id'] as String? ?? '';
    final imamConfirmed = widget.match['imam_confirmed'] as bool? ?? false;
    final familyIntroDone =
        widget.match['family_intro_done'] as bool? ?? false;
    final age = (widget.match['other_user_age'] as num?)?.toInt();
    final niyyah =
        NiyyahTypeExt.fromString(widget.match['other_user_niyyah'] as String?);
    final madhab = widget.match['other_user_madhab'] as String? ?? '';
    final city = widget.match['other_user_city'] as String? ?? '';

    final timerStr = widget.match['niyyah_timer_ends_at'] as String?;
    final timerEndsAt =
        timerStr != null ? DateTime.tryParse(timerStr) : null;
    final daysLeft = timerEndsAt?.difference(DateTime.now()).inDays;

    return GestureDetector(
      onTap: () => context.push('/chat/$matchId?userId=$otherUserId'),
      child: Container(
        padding: const EdgeInsets.all(AppSpacing.md),
        decoration: const BoxDecoration(
          color: AppColors.surface,
          borderRadius: AppRadius.card,
          boxShadow: AppShadows.soft,
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Stack(
                  children: [
                    CircleAvatar(
                      radius: 32,
                      backgroundColor: AppColors.surfaceVariant,
                      backgroundImage: avatarUrl != null
                          ? CachedNetworkImageProvider(avatarUrl)
                          : null,
                      child: avatarUrl == null
                          ? const Icon(Icons.person,
                              size: 32, color: AppColors.textHint)
                          : null,
                    ),
                    Positioned(
                      right: 0,
                      bottom: 0,
                      child: TrustScoreBadge(score: trustScore, size: 22),
                    ),
                  ],
                ),
                const SizedBox(width: AppSpacing.md),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        (age != null && age > 0) ? '$name, $age' : name,
                        style: GoogleFonts.nunito(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: AppColors.textPrimary,
                        ),
                      ),
                      if (city.isNotEmpty) ...[
                        const SizedBox(height: 2),
                        Row(
                          children: [
                            const Icon(Icons.location_on_outlined,
                                size: 12, color: AppColors.textHint),
                            const SizedBox(width: 3),
                            Text(
                              city,
                              style: GoogleFonts.nunito(
                                  fontSize: 12,
                                  color: AppColors.textSecondary),
                            ),
                          ],
                        ),
                      ],
                      const SizedBox(height: 4),
                      Wrap(
                        spacing: 4,
                        runSpacing: 4,
                        children: [
                          NiyyahBadge(niyyah: niyyah),
                          if (madhab.isNotEmpty) MadhabBadge(madhab: madhab),
                        ],
                      ),
                      const SizedBox(height: 4),
                      _statusRow(imamConfirmed, familyIntroDone),
                    ],
                  ),
                ),
                if (daysLeft != null && !imamConfirmed)
                  _TimerChip(daysLeft: daysLeft),
              ],
            ),
            if (!familyIntroDone && !imamConfirmed) ...[
              const SizedBox(height: AppSpacing.sm),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  onPressed:
                      _introLoading ? null : () => _doFamilyIntro(matchId),
                  style: OutlinedButton.styleFrom(
                    foregroundColor: AppColors.secondary,
                    side: const BorderSide(
                        color: AppColors.goldBorder, width: 1),
                    padding: const EdgeInsets.symmetric(vertical: 8),
                    textStyle:
                        GoogleFonts.nunito(fontWeight: FontWeight.w600),
                  ),
                  icon: _introLoading
                      ? const SizedBox(
                          width: 14,
                          height: 14,
                          child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: AppColors.secondary),
                        )
                      : const Icon(Icons.group_add_outlined, size: 16),
                  label: const Text('Отбасыны таныстыру'),
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _statusRow(bool imamConfirmed, bool familyIntroDone) {
    if (imamConfirmed) {
      return Row(
        children: [
          const Icon(Icons.favorite, size: 13, color: AppColors.primary),
          const SizedBox(width: 4),
          Text(
            'Никах расталды',
            style: GoogleFonts.nunito(
                fontSize: 12,
                color: AppColors.primary,
                fontWeight: FontWeight.w600),
          ),
        ],
      );
    }
    if (familyIntroDone) {
      return Row(
        children: [
          const Icon(Icons.group, size: 13, color: AppColors.secondary),
          const SizedBox(width: 4),
          Text(
            'Отбасы таныстырылды',
            style:
                GoogleFonts.nunito(fontSize: 12, color: AppColors.textSecondary),
          ),
        ],
      );
    }
    return Text(
      'Хабарлама жіберіңіз',
      style: GoogleFonts.nunito(fontSize: 12, color: AppColors.textSecondary),
    );
  }
}

class _TimerChip extends StatelessWidget {
  final int daysLeft;
  const _TimerChip({required this.daysLeft});

  @override
  Widget build(BuildContext context) {
    final isUrgent = daysLeft <= 7;
    final color = isUrgent ? AppColors.accent : AppColors.secondary;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.12),
        borderRadius: AppRadius.chip,
        border: Border.all(color: color.withValues(alpha: 0.4)),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            '$daysLeft',
            style: GoogleFonts.nunito(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: color,
            ),
          ),
          Text('күн', style: GoogleFonts.nunito(fontSize: 10, color: color)),
        ],
      ),
    );
  }
}
