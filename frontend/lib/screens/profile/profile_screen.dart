import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../core/constants/api_constants.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../models/profile.dart';
import '../../providers/auth_provider.dart';
import '../../widgets/halal_pattern_painter.dart';
import '../../widgets/niyyah_badge.dart';
import '../../widgets/trust_score_badge.dart';

final _ownProfileProvider = FutureProvider<Profile>((ref) async {
  final dio = ref.watch(dioClientProvider).dio;
  final resp = await dio.get(ApiConstants.profile);
  return Profile.fromJson(resp.data as Map<String, dynamic>);
});

class ProfileScreen extends ConsumerWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncProfile = ref.watch(_ownProfileProvider);

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
          'Менің профилім',
          style: GoogleFonts.nunito(
            fontSize: 20,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.edit_outlined, color: Colors.white),
            onPressed: () => context.push('/kyc'),
            tooltip: 'Өңдеу',
          ),
        ],
      ),
      body: asyncProfile.when(
        loading: () => const Center(
            child: CircularProgressIndicator(color: AppColors.primary)),
        error: (e, _) => _buildError(context, ref, e.toString()),
        data: (profile) => _buildProfile(context, ref, profile),
      ),
    );
  }

  Widget _buildProfile(
      BuildContext context, WidgetRef ref, Profile profile) {
    final niyyah = NiyyahTypeExt.fromString(profile.niyyah);

    return SingleChildScrollView(
      child: Column(
        children: [
          // Header with avatar
          _ProfileHeader(profile: profile),

          const SizedBox(height: AppSpacing.lg),

          // Trust Score section
          _InfoCard(
            child: Row(
              children: [
                AnimatedTrustScoreBadge(
                    score: profile.trustScore, size: 64),
                const SizedBox(width: AppSpacing.md),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Сенім ұпайы',
                        style: GoogleFonts.nunito(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: AppColors.textPrimary,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        _trustDescription(profile.trustScore),
                        style: GoogleFonts.nunito(
                            fontSize: 13,
                            color: AppColors.textSecondary,
                            height: 1.4),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),

          const SizedBox(height: AppSpacing.sm),

          // Niyyah + Madhab + Languages
          _InfoCard(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _SectionLabel('Ислами сәйкестік'),
                const SizedBox(height: AppSpacing.sm),
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: [
                    NiyyahBadge(niyyah: niyyah, large: true),
                    if (profile.madhab != null &&
                        profile.madhab!.isNotEmpty)
                      MadhabBadge(madhab: profile.madhab!),
                    ...profile.languages.map(
                      (l) => Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 10, vertical: 5),
                        decoration: BoxDecoration(
                          color: AppColors.surfaceVariant,
                          borderRadius: AppRadius.chip,
                        ),
                        child: Text(l,
                            style: GoogleFonts.nunito(
                                fontSize: 12,
                                color: AppColors.textSecondary)),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),

          const SizedBox(height: AppSpacing.sm),

          // Bio
          if (profile.bio != null && profile.bio!.isNotEmpty)
            _InfoCard(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _SectionLabel('Өзім туралы'),
                  const SizedBox(height: AppSpacing.sm),
                  Text(
                    profile.bio!,
                    style: GoogleFonts.nunito(
                        fontSize: 14,
                        color: AppColors.textPrimary,
                        height: 1.6),
                  ),
                ],
              ),
            ),

          const SizedBox(height: AppSpacing.sm),

          // KYC status
          _InfoCard(
            child: Row(
              children: [
                Icon(
                  profile.isKycVerified
                      ? Icons.verified_user
                      : Icons.shield_outlined,
                  color: profile.isKycVerified
                      ? AppColors.primary
                      : AppColors.textHint,
                  size: 28,
                ),
                const SizedBox(width: AppSpacing.md),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'KYC верификация',
                        style: GoogleFonts.nunito(
                          fontSize: 15,
                          fontWeight: FontWeight.w600,
                          color: AppColors.textPrimary,
                        ),
                      ),
                      Text(
                        profile.isKycVerified
                            ? 'Расталған'
                            : 'Расталмаған — Верификациядан өтіңіз',
                        style: GoogleFonts.nunito(
                            fontSize: 12,
                            color: profile.isKycVerified
                                ? AppColors.primary
                                : AppColors.textHint),
                      ),
                    ],
                  ),
                ),
                if (!profile.isKycVerified)
                  TextButton(
                    onPressed: () => context.push('/kyc'),
                    child: Text('Өту',
                        style: GoogleFonts.nunito(
                            color: AppColors.primary,
                            fontWeight: FontWeight.w600)),
                  ),
              ],
            ),
          ),

          const SizedBox(height: AppSpacing.lg),

          // Logout
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: AppSpacing.lg),
            child: OutlinedButton(
              onPressed: () async {
                await ref.read(authStateProvider.notifier).logout();
                if (context.mounted) context.go('/auth/login');
              },
              style: OutlinedButton.styleFrom(
                foregroundColor: AppColors.accent,
                side: const BorderSide(color: AppColors.accent),
              ),
              child: const Text('Шығу'),
            ),
          ),

          const SizedBox(height: AppSpacing.xxl),
        ],
      ),
    );
  }

  Widget _buildError(BuildContext context, WidgetRef ref, String msg) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.person_off, size: 48, color: AppColors.textHint),
          const SizedBox(height: 16),
          Text(msg,
              style: GoogleFonts.nunito(color: AppColors.textSecondary),
              textAlign: TextAlign.center),
          const SizedBox(height: 16),
          ElevatedButton(
            onPressed: () => ref.refresh(_ownProfileProvider.future),
            child: const Text('Қайтадан'),
          ),
        ],
      ),
    );
  }

  String _trustDescription(int score) {
    if (score >= 80) return 'Жоғары сенім деңгейі. Верификацияланған қауымдастық мүшесі.';
    if (score >= 60) return 'Жақсы сенім деңгейі. Белсенді қатысушы.';
    if (score >= 40) return 'Орташа сенім деңгейі. Белсендіңізді арттырыңыз.';
    return 'Бастапқы деңгей. KYC верификациядан өтіп ұпайды арттырыңыз.';
  }
}

// ─── Profile Header ───────────────────────────────────────────────────────────

class _ProfileHeader extends StatelessWidget {
  final Profile profile;
  const _ProfileHeader({required this.profile});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      color: AppColors.primaryDark,
      padding: const EdgeInsets.fromLTRB(
          AppSpacing.lg, AppSpacing.lg, AppSpacing.lg, AppSpacing.xl),
      child: Column(
        children: [
          Stack(
            children: [
              CircleAvatar(
                radius: 56,
                backgroundColor: AppColors.primaryLight,
                backgroundImage: profile.avatarUrl != null &&
                        profile.avatarUrl!.isNotEmpty &&
                        !profile.avatarBlurred
                    ? NetworkImage(profile.avatarUrl!)
                    : null,
                child: profile.avatarUrl == null ||
                        profile.avatarUrl!.isEmpty ||
                        profile.avatarBlurred
                    ? const Icon(Icons.person,
                        size: 56, color: AppColors.primary)
                    : null,
              ),
              if (profile.noPhotoMode)
                Positioned(
                  right: 0,
                  bottom: 0,
                  child: Container(
                    padding: const EdgeInsets.all(4),
                    decoration: const BoxDecoration(
                      color: AppColors.secondary,
                      shape: BoxShape.circle,
                    ),
                    child: const Icon(Icons.lock,
                        color: Colors.white, size: 14),
                  ),
                ),
            ],
          ),
          const SizedBox(height: AppSpacing.md),
          Text(
            profile.displayName,
            style: GoogleFonts.nunito(
              fontSize: 22,
              fontWeight: FontWeight.bold,
              color: Colors.white,
            ),
          ),
          if (profile.city != null) ...[
            const SizedBox(height: 4),
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.location_on_outlined,
                    size: 14, color: AppColors.secondary),
                const SizedBox(width: 4),
                Text(
                  profile.city!,
                  style: GoogleFonts.nunito(
                      fontSize: 14, color: AppColors.secondary),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }
}

// ─── Shared widgets ───────────────────────────────────────────────────────────

class _InfoCard extends StatelessWidget {
  final Widget child;
  const _InfoCard({required this.child});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.symmetric(horizontal: AppSpacing.md),
      padding: const EdgeInsets.all(AppSpacing.md),
      decoration: const BoxDecoration(
        color: AppColors.surface,
        borderRadius: AppRadius.card,
        boxShadow: AppShadows.soft,
      ),
      child: child,
    );
  }
}

class _SectionLabel extends StatelessWidget {
  final String text;
  const _SectionLabel(this.text);

  @override
  Widget build(BuildContext context) {
    return Row(
      children: [
        const IslamicStarWidget(size: 14, color: AppColors.secondary),
        const SizedBox(width: 6),
        Text(
          text,
          style: GoogleFonts.nunito(
            fontSize: 13,
            fontWeight: FontWeight.w600,
            color: AppColors.textSecondary,
          ),
        ),
      ],
    );
  }
}
