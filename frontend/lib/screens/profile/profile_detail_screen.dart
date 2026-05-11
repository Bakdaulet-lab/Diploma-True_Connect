import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:haptic_feedback/haptic_feedback.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../widgets/halal_pattern_painter.dart';
import '../../widgets/niyyah_badge.dart';
import '../../widgets/trust_score_badge.dart';
import '../../widgets/whisper_report_modal.dart';

class ProfileDetailScreen extends StatelessWidget {
  final Map<String, dynamic> profile;

  const ProfileDetailScreen({super.key, required this.profile});

  @override
  Widget build(BuildContext context) {
    final score = (profile['trustScore'] as num?)?.toInt() ?? 0;
    final niyyah =
        NiyyahTypeExt.fromString(profile['niyyah'] as String?);
    final madhab = profile['madhab'] as String? ?? '';
    final isKyc = profile['isKycVerified'] as bool? ?? false;

    return Scaffold(
      backgroundColor: AppColors.background,
      body: CustomScrollView(
        slivers: [
          _ProfileSliverAppBar(
            profile: profile,
            score: score,
            isKyc: isKyc,
            onBack: () {
              Haptics.vibrate(HapticsType.selection);
              Navigator.pop(context);
            },
          ),

          SliverToBoxAdapter(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Name / age / city header
                _ProfileHeader(profile: profile),

                const KazakhDivider(indent: AppSpacing.lg),
                const SizedBox(height: AppSpacing.md),

                // Niyyah + Madhab chips
                Padding(
                  padding: const EdgeInsets.symmetric(
                      horizontal: AppSpacing.lg),
                  child: Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: [
                      NiyyahBadge(niyyah: niyyah, large: true),
                      if (madhab.isNotEmpty) MadhabBadge(madhab: madhab),
                    ],
                  ),
                ),

                const SizedBox(height: AppSpacing.lg),

                // Trust Score section
                _TrustSection(score: score),

                const SizedBox(height: AppSpacing.md),
                const KazakhDivider(indent: AppSpacing.lg),
                const SizedBox(height: AppSpacing.md),

                // Bio
                if ((profile['bio'] as String?)?.isNotEmpty == true)
                  _BioSection(bio: profile['bio'] as String),

                // Prompts
                if ((profile['prompts'] as List?)?.isNotEmpty == true)
                  _PromptsSection(
                      prompts: profile['prompts'] as List),

                const SizedBox(height: 120),
              ],
            ),
          ),
        ],
      ),

      // Bottom action buttons — rectangular, not round FABs
      bottomNavigationBar: _BottomActions(
        onPass: () {
          Haptics.vibrate(HapticsType.medium);
          Navigator.pop(context, 'pass');
        },
        onLike: () {
          Haptics.vibrate(HapticsType.success);
          Navigator.pop(context, 'like');
        },
        onWhisper: () {
          final userId = profile['id'] as String? ?? '';
          if (userId.isNotEmpty) {
            showWhisperModal(
              context,
              matchId: profile['matchId'] as String? ?? '',
              reportedUserId: userId,
            );
          }
        },
      ),
    );
  }
}

// ─── Sliver App Bar with hero photo ──────────────────────────────────────────

class _ProfileSliverAppBar extends StatelessWidget {
  final Map<String, dynamic> profile;
  final int score;
  final bool isKyc;
  final VoidCallback onBack;

  const _ProfileSliverAppBar({
    required this.profile,
    required this.score,
    required this.isKyc,
    required this.onBack,
  });

  @override
  Widget build(BuildContext context) {
    return SliverAppBar(
      expandedHeight: MediaQuery.of(context).size.height * 0.52,
      pinned: true,
      backgroundColor: AppColors.primaryDark,
      leading: GestureDetector(
        onTap: onBack,
        child: Container(
          margin: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: Colors.black38,
            borderRadius: BorderRadius.circular(12),
          ),
          child: const Icon(Icons.arrow_back_ios_new,
              color: Colors.white, size: 18),
        ),
      ),
      flexibleSpace: FlexibleSpaceBar(
        background: Stack(
          fit: StackFit.expand,
          children: [
            // Hero photo
            Hero(
              tag: 'profile-img-${profile["id"]}',
              child: Image.network(
                profile['imageUrl'] as String? ?? '',
                fit: BoxFit.cover,
                errorBuilder: (_, __, ___) => Container(
                  color: AppColors.surfaceVariant,
                  child: const Icon(Icons.person,
                      size: 120, color: AppColors.textHint),
                ),
              ),
            ),

            // Bottom gradient
            Positioned.fill(
              child: DecoratedBox(
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                    stops: const [0.5, 1.0],
                    colors: [
                      Colors.transparent,
                      Colors.black.withValues(alpha: 0.6),
                    ],
                  ),
                ),
              ),
            ),

            // Trust score — top right
            Positioned(
              top: 60,
              right: 16,
              child: AnimatedTrustScoreBadge(score: score, size: 52),
            ),

            // KYC badge
            if (isKyc)
              Positioned(
                bottom: 16,
                left: 16,
                child: Container(
                  padding: const EdgeInsets.symmetric(
                      horizontal: 12, vertical: 6),
                  decoration: BoxDecoration(
                    color: AppColors.primary.withValues(alpha: 0.92),
                    borderRadius: AppRadius.chip,
                  ),
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.verified,
                          color: Colors.white, size: 14),
                      const SizedBox(width: 5),
                      Text(
                        'Верификацияланған',
                        style: GoogleFonts.nunito(
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                          color: Colors.white,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

// ─── Profile header: name + city ─────────────────────────────────────────────

class _ProfileHeader extends StatelessWidget {
  final Map<String, dynamic> profile;

  const _ProfileHeader({required this.profile});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(
          AppSpacing.lg, AppSpacing.lg, AppSpacing.lg, AppSpacing.md),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            '${profile["name"]}, ${profile["age"]}',
            style: GoogleFonts.nunito(
              fontSize: 26,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
              letterSpacing: -0.3,
            ),
          ),
          const SizedBox(height: 6),
          Row(
            children: [
              const Icon(Icons.location_on_outlined,
                  size: 15, color: AppColors.textHint),
              const SizedBox(width: 4),
              Text(
                profile['city'] as String? ?? '',
                style: GoogleFonts.nunito(
                  fontSize: 14,
                  color: AppColors.textSecondary,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

// ─── Trust score display ──────────────────────────────────────────────────────

class _TrustSection extends StatelessWidget {
  final int score;

  const _TrustSection({required this.score});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg),
      child: Container(
        padding: const EdgeInsets.all(AppSpacing.md),
        decoration: const BoxDecoration(
          color: AppColors.surface,
          borderRadius: AppRadius.card,
          boxShadow: AppShadows.soft,
        ),
        child: Row(
          children: [
            TrustScoreBadge(score: score, size: 52),
            const SizedBox(width: AppSpacing.md),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Сенім рейтингі',
                    style: GoogleFonts.nunito(
                      fontSize: 15,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  const SizedBox(height: 3),
                  Text(
                    'KYC верификация мен нақты кездесулерге негізделген',
                    style: GoogleFonts.nunito(
                      fontSize: 12,
                      color: AppColors.textSecondary,
                      height: 1.4,
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
}

// ─── Bio section ─────────────────────────────────────────────────────────────

class _BioSection extends StatefulWidget {
  final String bio;

  const _BioSection({required this.bio});

  @override
  State<_BioSection> createState() => _BioSectionState();
}

class _BioSectionState extends State<_BioSection> {
  bool _expanded = false;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(
          AppSpacing.lg, 0, AppSpacing.lg, AppSpacing.md),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const IslamicStarWidget(size: 14, color: AppColors.secondary),
              const SizedBox(width: 8),
              Text(
                'Өзі туралы',
                style: GoogleFonts.nunito(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Text(
            widget.bio,
            style: GoogleFonts.nunito(
              fontSize: 14,
              color: AppColors.textSecondary,
              height: 1.6,
            ),
            maxLines: _expanded ? null : 3,
            overflow: _expanded ? null : TextOverflow.ellipsis,
          ),
          if (widget.bio.length > 120)
            GestureDetector(
              onTap: () => setState(() => _expanded = !_expanded),
              child: Padding(
                padding: const EdgeInsets.only(top: 4),
                child: Text(
                  _expanded ? 'Жасыру' : 'Толығырақ оқу',
                  style: GoogleFonts.nunito(
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                    color: AppColors.primary,
                  ),
                ),
              ),
            ),
          const SizedBox(height: AppSpacing.md),
          const KazakhDivider(indent: 0),
        ],
      ),
    );
  }
}

// ─── Prompts section ──────────────────────────────────────────────────────────

class _PromptsSection extends StatelessWidget {
  final List prompts;

  const _PromptsSection({required this.prompts});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const SizedBox(height: AppSpacing.md),
          Row(
            children: [
              const IslamicStarWidget(size: 14, color: AppColors.secondary),
              const SizedBox(width: 8),
              Text(
                'Сұрақтар',
                style: GoogleFonts.nunito(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          ...prompts.map((p) {
            final prompt = p as Map;
            return Container(
              width: double.infinity,
              margin: const EdgeInsets.only(bottom: 12),
              padding: const EdgeInsets.all(AppSpacing.md),
              decoration: const BoxDecoration(
                color: AppColors.surface,
                borderRadius: AppRadius.card,
                boxShadow: AppShadows.soft,
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    padding: const EdgeInsets.symmetric(
                        horizontal: 8, vertical: 3),
                    decoration: const BoxDecoration(
                      color: AppColors.primaryLight,
                      borderRadius: AppRadius.chip,
                    ),
                    child: Text(
                      prompt['question'] as String? ?? '',
                      style: GoogleFonts.nunito(
                        fontSize: 12,
                        color: AppColors.primary,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    prompt['answer'] as String? ?? '',
                    style: GoogleFonts.nunito(
                      fontSize: 14,
                      color: AppColors.textPrimary,
                      height: 1.5,
                    ),
                  ),
                ],
              ),
            );
          }),
        ],
      ),
    );
  }
}

// ─── Bottom action bar ────────────────────────────────────────────────────────

class _BottomActions extends StatelessWidget {
  final VoidCallback onPass;
  final VoidCallback onLike;
  final VoidCallback onWhisper;

  const _BottomActions({
    required this.onPass,
    required this.onLike,
    required this.onWhisper,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.fromLTRB(
        AppSpacing.md,
        AppSpacing.md,
        AppSpacing.md,
        AppSpacing.md + MediaQuery.of(context).padding.bottom,
      ),
      decoration: const BoxDecoration(
        color: AppColors.surface,
        boxShadow: [
          BoxShadow(
            color: Color(0x14000000),
            blurRadius: 12,
            offset: Offset(0, -4),
          ),
        ],
      ),
      child: Row(
        children: [
          // Pass — outlined
          Expanded(
            child: OutlinedButton(
              onPressed: onPass,
              style: OutlinedButton.styleFrom(
                minimumSize: const Size(0, 52),
                side: const BorderSide(
                    color: AppColors.surfaceVariant, width: 1.5),
                shape: const RoundedRectangleBorder(
                    borderRadius: AppRadius.button),
              ),
              child: Text(
                '✕  Өткізу',
                style: GoogleFonts.nunito(
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                  color: AppColors.textSecondary,
                ),
              ),
            ),
          ),

          const SizedBox(width: AppSpacing.sm),

          // Like — filled primary
          Expanded(
            flex: 2,
            child: ElevatedButton(
              onPressed: onLike,
              style: ElevatedButton.styleFrom(
                minimumSize: const Size(0, 52),
                backgroundColor: AppColors.primary,
                shape: const RoundedRectangleBorder(
                    borderRadius: AppRadius.button),
              ),
              child: Text(
                '♥  Ұнайды',
                style: GoogleFonts.nunito(
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                  color: Colors.white,
                ),
              ),
            ),
          ),

          const SizedBox(width: AppSpacing.sm),

          // Whisper ghost report button
          SizedBox(
            width: 52,
            height: 52,
            child: OutlinedButton(
              onPressed: onWhisper,
              style: OutlinedButton.styleFrom(
                padding: EdgeInsets.zero,
                side: const BorderSide(
                    color: AppColors.surfaceVariant, width: 1.5),
                shape: const RoundedRectangleBorder(
                    borderRadius: AppRadius.button),
              ),
              child: const Icon(
                Icons.report_gmailerrorred_outlined,
                color: AppColors.textHint,
                size: 20,
              ),
            ),
          ),
        ],
      ),
    );
  }
}
