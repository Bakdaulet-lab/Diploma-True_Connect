import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../providers/matching_provider.dart';
import '../../widgets/halal_pattern_painter.dart';
import '../../widgets/trust_score_badge.dart';

class MatchesScreen extends ConsumerWidget {
  const MatchesScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncMatches = ref.watch(matchesListProvider);

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.primaryDark,
        centerTitle: true,
        bottom: const PreferredSize(
          preferredSize: Size.fromHeight(1),
          child:
              Divider(height: 1, thickness: 1, color: AppColors.goldBorder),
        ),
        title: Text(
          'Сәйкестіктер',
          style: GoogleFonts.nunito(
            fontSize: 20,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
        ),
      ),
      body: asyncMatches.when(
        loading: () =>
            const Center(child: CircularProgressIndicator(color: AppColors.primary)),
        error: (e, _) => _buildError(context, ref, e.toString()),
        data: (matches) => matches.isEmpty
            ? _buildEmpty()
            : RefreshIndicator(
                color: AppColors.primary,
                onRefresh: () => ref.refresh(matchesListProvider.future),
                child: ListView.separated(
                  padding: const EdgeInsets.all(AppSpacing.md),
                  itemCount: matches.length,
                  separatorBuilder: (_, __) =>
                      const SizedBox(height: AppSpacing.sm),
                  itemBuilder: (context, index) =>
                      _MatchCard(match: matches[index]),
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
            style: GoogleFonts.nunito(
                fontSize: 14, color: AppColors.textSecondary),
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

// ─── Match Card ───────────────────────────────────────────────────────────────

class _MatchCard extends StatelessWidget {
  final Map<String, dynamic> match;
  const _MatchCard({required this.match});

  @override
  Widget build(BuildContext context) {
    final name = match['other_user_name'] as String? ?? '';
    final avatarUrl = match['other_user_avatar_url'] as String?;
    final trustScore =
        (match['other_user_trust_score'] as num?)?.toInt() ?? 0;
    final matchId = match['id'] as String? ?? '';
    final imamConfirmed = match['imam_confirmed'] as bool? ?? false;
    final familyIntroDone = match['family_intro_done'] as bool? ?? false;

    // Niyyah timer
    final timerStr = match['niyyah_timer_ends_at'] as String?;
    final timerEndsAt =
        timerStr != null ? DateTime.tryParse(timerStr) : null;
    final daysLeft = timerEndsAt?.difference(DateTime.now()).inDays;

    return GestureDetector(
      onTap: () => context.push('/chat/$matchId'),
      child: Container(
        padding: const EdgeInsets.all(AppSpacing.md),
        decoration: const BoxDecoration(
          color: AppColors.surface,
          borderRadius: AppRadius.card,
          boxShadow: AppShadows.soft,
        ),
        child: Row(
          children: [
            // Avatar
            Stack(
              children: [
                CircleAvatar(
                  radius: 32,
                  backgroundColor: AppColors.surfaceVariant,
                  backgroundImage:
                      avatarUrl != null ? NetworkImage(avatarUrl) : null,
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

            // Name + status
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    name,
                    style: GoogleFonts.nunito(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  const SizedBox(height: 4),
                  _statusRow(imamConfirmed, familyIntroDone),
                ],
              ),
            ),

            // Niyyah timer
            if (daysLeft != null && !imamConfirmed)
              _TimerChip(daysLeft: daysLeft),
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
            style: GoogleFonts.nunito(
                fontSize: 12, color: AppColors.textSecondary),
          ),
        ],
      );
    }
    return Text(
      'Хабарлама жіберіңіз',
      style:
          GoogleFonts.nunito(fontSize: 12, color: AppColors.textSecondary),
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
          Text(
            'күн',
            style: GoogleFonts.nunito(fontSize: 10, color: color),
          ),
        ],
      ),
    );
  }
}
