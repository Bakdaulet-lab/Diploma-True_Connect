import 'package:cached_network_image/cached_network_image.dart';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:haptic_feedback/haptic_feedback.dart';
import '../../core/constants/api_constants.dart';
import '../../core/network/dio_error_message.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../models/profile.dart';
import '../../providers/auth_provider.dart';
import '../../widgets/halal_pattern_painter.dart';
import '../../widgets/niyyah_badge.dart';
import '../../widgets/trust_score_badge.dart';
import '../../widgets/whisper_report_modal.dart';

final publicProfileProvider =
    FutureProvider.family<Profile, String>((ref, userId) async {
  if (userId.isEmpty) {
    throw 'profile id is missing';
  }

  final dio = ref.watch(dioClientProvider).dio;
  try {
    final resp = await dio.get('${ApiConstants.profiles}/$userId');
    final payload = resp.data is Map
        ? Map<String, dynamic>.from(resp.data as Map)
        : <String, dynamic>{};
    return Profile.fromJson(payload);
  } on DioException catch (e) {
    throw dioErrorMessage(e);
  }
});

Map<String, dynamic> _legacyProfileMap(Profile profile) {
  return {
    'id': profile.userId,
    'name': profile.displayName,
    'displayName': profile.displayName,
    'age': profile.age,
    'city': profile.city ?? '',
    'bio': profile.bio ?? '',
    'imageUrl': profile.avatarUrl ?? '',
    'avatarUrl': profile.avatarUrl ?? '',
    'trustScore': profile.trustScore,
    'isKycVerified': profile.isKycVerified,
    'niyyah': profile.niyyah,
    'madhab': profile.madhab,
    'prompts': profile.prompts,
    'matchId': '',
    'languages': profile.languages,
    'noPhotoMode': profile.noPhotoMode,
    'maritalStatus': profile.maritalStatus,
  };
}

class ProfileDetailScreen extends ConsumerWidget {
  final Map<String, dynamic> profile;

  const ProfileDetailScreen({super.key, required this.profile});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final userId = profile['id'] as String? ?? '';
    if (userId.isEmpty) {
      return const Scaffold(
        backgroundColor: AppColors.background,
        body: Center(
          child: Text('Профиль ID жоқ'),
        ),
      );
    }

    final asyncProfile = ref.watch(publicProfileProvider(userId));

    return asyncProfile.when(
      loading: () => const Scaffold(
        backgroundColor: AppColors.background,
        body: Center(
          child: CircularProgressIndicator(color: AppColors.primary),
        ),
      ),
      error: (error, _) => Scaffold(
        backgroundColor: AppColors.background,
        body: Center(
          child: Padding(
            padding: const EdgeInsets.all(AppSpacing.lg),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(Icons.wifi_off,
                    size: 48, color: AppColors.textHint),
                const SizedBox(height: 16),
                Text(
                  error.toString(),
                  style: GoogleFonts.nunito(color: AppColors.textSecondary),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 16),
                ElevatedButton(
                  onPressed: () => ref.refresh(publicProfileProvider(userId)),
                  child: const Text('Қайтадан көру'),
                ),
              ],
            ),
          ),
        ),
      ),
      data: (loadedProfile) => _ProfileDetailView(
        profile: _legacyProfileMap(loadedProfile),
      ),
    );
  }
}

class _ProfileDetailView extends StatelessWidget {
  final Map<String, dynamic> profile;

  const _ProfileDetailView({required this.profile});

  @override
  Widget build(BuildContext context) {
    final score = (profile['trustScore'] as num?)?.toInt() ?? 0;
    final niyyah = NiyyahTypeExt.fromString(profile['niyyah'] as String?);
    final madhab = profile['madhab'] as String? ?? '';
    final isKyc = profile['isKycVerified'] as bool? ?? false;
    final matchId = profile['matchId'] as String? ?? '';

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
                _ProfileHeader(profile: profile),
                const KazakhDivider(indent: AppSpacing.lg),
                const SizedBox(height: AppSpacing.md),
                Padding(
                  padding:
                      const EdgeInsets.symmetric(horizontal: AppSpacing.lg),
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
                _TrustSection(score: score),
                const SizedBox(height: AppSpacing.md),
                const KazakhDivider(indent: AppSpacing.lg),
                const SizedBox(height: AppSpacing.md),
                if ((profile['bio'] as String?)?.isNotEmpty == true)
                  _BioSection(bio: profile['bio'] as String),
                if ((profile['prompts'] as List?)?.isNotEmpty == true)
                  _PromptsSection(prompts: profile['prompts'] as List),
                const SizedBox(height: 120),
              ],
            ),
          ),
        ],
      ),
      bottomNavigationBar: _BottomActions(
        onPass: () {
          Haptics.vibrate(HapticsType.medium);
          Navigator.pop(context, 'pass');
        },
        onLike: () {
          Haptics.vibrate(HapticsType.success);
          Navigator.pop(context, 'like');
        },
        onWhisper: matchId.isNotEmpty
            ? () {
                final userId = profile['id'] as String? ?? '';
                if (userId.isNotEmpty) {
                  showWhisperModal(
                    context,
                    matchId: matchId,
                    reportedUserId: userId,
                  );
                }
              }
            : null,
      ),
    );
  }
}

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
    final avatarUrl = (profile['imageUrl'] as String?)?.isNotEmpty == true
        ? profile['imageUrl'] as String?
        : (profile['avatarUrl'] as String?)?.isNotEmpty == true
            ? profile['avatarUrl'] as String?
            : null;

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
            Hero(
              tag: 'profile-img-${profile["id"]}',
              child: avatarUrl != null
                  ? CachedNetworkImage(
                      imageUrl: avatarUrl,
                      fit: BoxFit.cover,
                      errorWidget: (_, __, ___) => Container(
                        color: AppColors.surfaceVariant,
                        child: const Icon(Icons.person,
                            size: 120, color: AppColors.textHint),
                      ),
                    )
                  : Container(
                      color: AppColors.surfaceVariant,
                      child: const Icon(Icons.person,
                          size: 120, color: AppColors.textHint),
                    ),
            ),
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
            Positioned(
              top: 60,
              right: 16,
              child: AnimatedTrustScoreBadge(score: score, size: 52),
            ),
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

class _ProfileHeader extends StatelessWidget {
  final Map<String, dynamic> profile;

  const _ProfileHeader({required this.profile});

  @override
  Widget build(BuildContext context) {
    final name = profile['name'] as String? ?? '';
    final age = profile['age'];
    final title = age is num ? '$name, ${age.toInt()}' : name;

    return Padding(
      padding: const EdgeInsets.fromLTRB(
          AppSpacing.lg, AppSpacing.lg, AppSpacing.lg, AppSpacing.md),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
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
              const Icon(Icons.star_border,
                  size: 14, color: AppColors.secondary),
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
              const Icon(Icons.star_border,
                  size: 14, color: AppColors.secondary),
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
          ...prompts.map((item) {
            final prompt = item is Map
                ? Map<String, dynamic>.from(item)
                : <String, dynamic>{};
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

class _BottomActions extends StatelessWidget {
  final VoidCallback onPass;
  final VoidCallback onLike;
  final VoidCallback? onWhisper;

  const _BottomActions({
    required this.onPass,
    required this.onLike,
    this.onWhisper,
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
