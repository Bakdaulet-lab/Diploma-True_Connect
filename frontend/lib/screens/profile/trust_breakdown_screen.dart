import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../models/trust_breakdown.dart';
import '../../providers/trust_provider.dart';
import '../../widgets/common_widgets.dart';
import '../../widgets/trust_score_badge.dart';

class TrustBreakdownScreen extends ConsumerWidget {
  const TrustBreakdownScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final state = ref.watch(trustBreakdownProvider);

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        title: Text(
          'Сенім ұпайы',
          style: GoogleFonts.nunito(
            fontSize: 18,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
        ),
      ),
      body: state.when(
        loading: () => const LoadingWidget(message: 'Жүктелуде...'),
        error: (e, _) => ErrorRetryWidget(
          message: e.toString(),
          onRetry: () => ref.invalidate(trustBreakdownProvider),
        ),
        data: (b) => _Body(b: b),
      ),
    );
  }
}

class _Body extends StatelessWidget {
  const _Body({required this.b});
  final TrustBreakdown b;

  @override
  Widget build(BuildContext context) {
    return ListView(
      padding: const EdgeInsets.all(AppSpacing.lg),
      children: [
        Center(
          child: Column(
            children: [
              AnimatedTrustScoreBadge(score: b.score, size: 88),
              const SizedBox(height: AppSpacing.sm),
              Text(
                b.badge,
                style: GoogleFonts.nunito(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: AppSpacing.lg),
        Text(
          'Ұпай қалай есептеледі',
          style: GoogleFonts.nunito(
            fontSize: 16,
            fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
          ),
        ),
        const SizedBox(height: AppSpacing.sm),
        _ComponentCard(
          rows: [
            _Row(
              label: 'Орташа баға (${b.ratingCount} баға)',
              value: '${b.smoothedRating.toStringAsFixed(2)} / 5',
              detail: '→ ${b.baseScore.toStringAsFixed(0)} ұпай',
            ),
            _Row(
              label: 'KYC верификация бонусы',
              value: '+${b.kycBonus.toStringAsFixed(0)}',
              positive: true,
            ),
            _Row(
              label: 'Шағымдар айыппұлы (${b.reportCount})',
              value: '−${b.reportPenalty.toStringAsFixed(0)}',
              negative: true,
            ),
            _Row(
              label: 'Қорытынды ұпай',
              value: '${b.score}',
              bold: true,
            ),
          ],
        ),
        const SizedBox(height: AppSpacing.md),
        Text(
          'Баға бейтарап 2.5/5 деңгейіне қарай тегістеледі: жаңа '
          'қолданушылар әділ 50 ұпайдан бастайды. Сенімді, расталған '
          'қолданушылардың бағалары көбірек салмаққа ие.',
          style: GoogleFonts.nunito(
            fontSize: 13,
            color: AppColors.textSecondary,
            height: 1.4,
          ),
        ),
        if (b.recentRatings.isNotEmpty) ...[
          const SizedBox(height: AppSpacing.lg),
          Text(
            'Соңғы бағалар',
            style: GoogleFonts.nunito(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: AppSpacing.sm),
          ...b.recentRatings.map((r) => _RatingTile(r: r)),
        ],
      ],
    );
  }
}

class _ComponentCard extends StatelessWidget {
  const _ComponentCard({required this.rows});
  final List<Widget> rows;

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(AppSpacing.md),
      decoration: const BoxDecoration(
        color: AppColors.surface,
        borderRadius: AppRadius.card,
      ),
      child: Column(children: rows),
    );
  }
}

class _Row extends StatelessWidget {
  const _Row({
    required this.label,
    required this.value,
    this.detail,
    this.positive = false,
    this.negative = false,
    this.bold = false,
  });

  final String label;
  final String value;
  final String? detail;
  final bool positive;
  final bool negative;
  final bool bold;

  @override
  Widget build(BuildContext context) {
    final Color valueColor = positive
        ? AppColors.trustGood
        : negative
            ? AppColors.accent
            : AppColors.textPrimary;
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        children: [
          Expanded(
            child: Text(
              label,
              style: GoogleFonts.nunito(
                fontSize: 14,
                fontWeight: bold ? FontWeight.bold : FontWeight.w500,
                color: AppColors.textPrimary,
              ),
            ),
          ),
          if (detail != null) ...[
            Text(
              detail!,
              style: GoogleFonts.nunito(
                fontSize: 12,
                color: AppColors.textSecondary,
              ),
            ),
            const SizedBox(width: 8),
          ],
          Text(
            value,
            style: GoogleFonts.nunito(
              fontSize: 15,
              fontWeight: FontWeight.bold,
              color: valueColor,
            ),
          ),
        ],
      ),
    );
  }
}

class _RatingTile extends StatelessWidget {
  const _RatingTile({required this.r});
  final TrustRating r;

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: AppSpacing.sm),
      padding: const EdgeInsets.all(AppSpacing.md),
      decoration: const BoxDecoration(
        color: AppColors.surface,
        borderRadius: AppRadius.card,
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            '${r.rating}★',
            style: GoogleFonts.nunito(
              fontSize: 15,
              fontWeight: FontWeight.bold,
              color: AppColors.secondary,
            ),
          ),
          const SizedBox(width: AppSpacing.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  r.context,
                  style: GoogleFonts.nunito(
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                    color: AppColors.textSecondary,
                  ),
                ),
                if (r.comment.isNotEmpty) ...[
                  const SizedBox(height: 2),
                  Text(
                    r.comment,
                    style: GoogleFonts.nunito(
                      fontSize: 13,
                      color: AppColors.textPrimary,
                      height: 1.3,
                    ),
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }
}
