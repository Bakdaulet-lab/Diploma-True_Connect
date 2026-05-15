import 'package:cached_network_image/cached_network_image.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:dio/dio.dart';
import 'package:image_picker/image_picker.dart';
import '../../core/constants/api_constants.dart';
import '../../core/network/dio_error_message.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../models/profile.dart';
import '../../providers/auth_provider.dart';
import '../../providers/profile_provider.dart';
import '../../widgets/halal_pattern_painter.dart';
import '../../widgets/niyyah_badge.dart';
import '../../widgets/trust_score_badge.dart';

class ProfileScreen extends ConsumerWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncProfile = ref.watch(ownProfileProvider);

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
            onPressed: () => context.push('/profile/edit'),
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

          // Completeness progress bar
          _ProfileCompletenessBar(profile: profile),

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
                const _SectionLabel('Ислами сәйкестік'),
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
                        decoration: const BoxDecoration(
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
                  const _SectionLabel('Өзім туралы'),
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
    final isMissingProfile =
        msg.contains('NOT_FOUND') || msg.contains('resource not found');

    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.lg),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              isMissingProfile ? Icons.person_add_alt_1 : Icons.person_off,
              size: 48,
              color: AppColors.textHint,
            ),
            const SizedBox(height: 16),
            Text(
              isMissingProfile ? 'Профиль әлі жасалмаған' : msg,
              style: GoogleFonts.nunito(color: AppColors.textSecondary),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              isMissingProfile
                  ? 'Профильді бір рет толтырсаңыз, бұл бет қалыпты жұмыс істейді.'
                  : 'Профильді қайтадан жүктеп көріңіз.',
              style: GoogleFonts.nunito(
                  fontSize: 13, color: AppColors.textHint),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 20),
            if (isMissingProfile) ...[
              ElevatedButton(
                onPressed: () async {
                  final created = await _showCreateProfileSheet(context);
                  if (created == true) {
                    ref.invalidate(ownProfileProvider);
                  }
                },
                child: const Text('Профиль жасау'),
              ),
              const SizedBox(height: 12),
            ],
            ElevatedButton(
              onPressed: () => ref.refresh(ownProfileProvider.future),
              child: const Text('Қайтадан'),
            ),
          ],
        ),
      ),
    );
  }

  Future<bool?> _showCreateProfileSheet(BuildContext context) {
    return showModalBottomSheet<bool>(
      context: context,
      isScrollControlled: true,
      backgroundColor: AppColors.surface,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (_) => const _CreateProfileSheet(),
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

class _ProfileHeader extends ConsumerStatefulWidget {
  final Profile profile;
  const _ProfileHeader({required this.profile});

  @override
  ConsumerState<_ProfileHeader> createState() => _ProfileHeaderState();
}

class _ProfileHeaderState extends ConsumerState<_ProfileHeader> {
  bool _uploading = false;

  Future<void> _pickAndUploadAvatar() async {
    final picker = ImagePicker();
    final picked = await picker.pickImage(
      source: ImageSource.gallery,
      maxWidth: 1024,
      maxHeight: 1024,
      imageQuality: 85,
    );
    if (picked == null || !mounted) return;

    setState(() => _uploading = true);
    try {
      final dio = ref.read(dioClientProvider).dio;
      final bytes = await picked.readAsBytes();
      final formData = FormData.fromMap({
        'photo': MultipartFile.fromBytes(
          bytes,
          filename: picked.name,
          contentType: DioMediaType('image', 'jpeg'),
        ),
      });
      await dio.post(ApiConstants.profilePhoto, data: formData);
      ref.invalidate(ownProfileProvider);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(
            content: Text('Фото жүктелді'),
            backgroundColor: AppColors.primary,
          ),
        );
      }
    } on DioException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(dioErrorMessage(e))),
        );
      }
    } finally {
      if (mounted) setState(() => _uploading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final profile = widget.profile;
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
                    ? CachedNetworkImageProvider(profile.avatarUrl!)
                    : null,
                child: profile.avatarUrl == null ||
                        profile.avatarUrl!.isEmpty ||
                        profile.avatarBlurred
                    ? const Icon(Icons.person,
                        size: 56, color: AppColors.primary)
                    : null,
              ),
              // Camera button to upload avatar
              Positioned(
                right: 0,
                bottom: 0,
                child: GestureDetector(
                  onTap: _uploading ? null : _pickAndUploadAvatar,
                  child: Container(
                    padding: const EdgeInsets.all(6),
                    decoration: BoxDecoration(
                      color: _uploading ? AppColors.textHint : AppColors.secondary,
                      shape: BoxShape.circle,
                      border: Border.all(color: AppColors.primaryDark, width: 2),
                    ),
                    child: _uploading
                        ? const SizedBox(
                            width: 14,
                            height: 14,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: Colors.white,
                            ),
                          )
                        : const Icon(Icons.camera_alt,
                            color: Colors.white, size: 14),
                  ),
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

class _CreateProfileSheet extends ConsumerStatefulWidget {
  const _CreateProfileSheet();

  @override
  ConsumerState<_CreateProfileSheet> createState() =>
      _CreateProfileSheetState();
}

class _CreateProfileSheetState extends ConsumerState<_CreateProfileSheet> {
  final _formKey = GlobalKey<FormState>();
  final _displayNameCtr = TextEditingController();
  final _cityCtr = TextEditingController();
  final _bioCtr = TextEditingController();
  bool _saving = false;
  String? _error;

  @override
  void dispose() {
    _displayNameCtr.dispose();
    _cityCtr.dispose();
    _bioCtr.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!(_formKey.currentState?.validate() ?? false)) return;

    setState(() {
      _saving = true;
      _error = null;
    });

    try {
      final dio = ref.read(dioClientProvider).dio;
      await dio.put(ApiConstants.profile, data: {
        'display_name': _displayNameCtr.text.trim(),
        if (_cityCtr.text.trim().isNotEmpty) 'city': _cityCtr.text.trim(),
        if (_bioCtr.text.trim().isNotEmpty) 'bio': _bioCtr.text.trim(),
      });

      if (!mounted) return;
      Navigator.pop(context, true);
    } on DioException catch (e) {
      if (!mounted) return;
      setState(() {
        _saving = false;
        _error = dioErrorMessage(e);
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _saving = false;
        _error = e.toString();
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return SafeArea(
      child: Padding(
        padding: EdgeInsets.fromLTRB(
          AppSpacing.lg,
          AppSpacing.lg,
          AppSpacing.lg,
          AppSpacing.lg + MediaQuery.of(context).viewInsets.bottom,
        ),
        child: SingleChildScrollView(
          child: Form(
            key: _formKey,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    const Icon(Icons.person_add_alt_1,
                        color: AppColors.primary, size: 22),
                    const SizedBox(width: 8),
                    Text(
                      'Профиль жасау',
                      style: GoogleFonts.nunito(
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                        color: AppColors.textPrimary,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Text(
                  'Кемінде аты-жөнін енгізіңіз, кейін профильді толықтыра аласыз.',
                  style: GoogleFonts.nunito(
                    fontSize: 13,
                    color: AppColors.textSecondary,
                    height: 1.4,
                  ),
                ),
                const SizedBox(height: AppSpacing.lg),
                TextFormField(
                  controller: _displayNameCtr,
                  textCapitalization: TextCapitalization.words,
                  decoration: const InputDecoration(
                    hintText: 'Аты-жөні',
                    prefixIcon:
                        Icon(Icons.badge_outlined, color: AppColors.textHint),
                  ),
                  validator: (value) => (value == null || value.trim().isEmpty)
                      ? 'Аты-жөніңізді енгізіңіз'
                      : null,
                ),
                const SizedBox(height: AppSpacing.md),
                TextFormField(
                  controller: _cityCtr,
                  textCapitalization: TextCapitalization.words,
                  decoration: const InputDecoration(
                    hintText: 'Қала (қосымша)',
                    prefixIcon: Icon(Icons.location_city_outlined,
                        color: AppColors.textHint),
                  ),
                ),
                const SizedBox(height: AppSpacing.md),
                TextFormField(
                  controller: _bioCtr,
                  maxLines: 3,
                  maxLength: 500,
                  decoration: const InputDecoration(
                    hintText: 'Өзіңіз туралы қысқаша (қосымша)',
                    prefixIcon:
                        Icon(Icons.edit_outlined, color: AppColors.textHint),
                  ),
                ),
                if (_error != null) ...[
                  const SizedBox(height: 8),
                  Text(
                    _error!,
                    style: GoogleFonts.nunito(
                      fontSize: 13,
                      color: AppColors.accent,
                    ),
                  ),
                ],
                const SizedBox(height: AppSpacing.lg),
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton(
                    onPressed: _saving ? null : _submit,
                    child: _saving
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(
                              color: Colors.white,
                              strokeWidth: 2,
                            ),
                          )
                        : const Text('Жасау'),
                  ),
                ),
                const SizedBox(height: AppSpacing.sm),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

// ─── Profile Completeness Bar ─────────────────────────────────────────────────

class _ProfileCompletenessBar extends StatelessWidget {
  final Profile profile;
  const _ProfileCompletenessBar({required this.profile});

  // Returns (filledCount, totalCount, firstMissingHint)
  (int, int, String?) _compute() {
    final checks = [
      (profile.displayName.isNotEmpty,         'Аты-жөн'),
      (profile.avatarUrl?.isNotEmpty == true,  'Фото'),
      (profile.bio?.isNotEmpty == true,        'Қысқаша таныстыру'),
      (profile.gender?.isNotEmpty == true,     'Жыныс'),
      (profile.niyyah?.isNotEmpty == true,     'Ниет'),
      (profile.city?.isNotEmpty == true,       'Қала'),
      (profile.madhab?.isNotEmpty == true,     'Мазхаб'),
      (profile.languages.isNotEmpty,           'Тілдер'),
    ];
    final filled = checks.where((c) => c.$1).length;
    final hint = checks.firstWhere((c) => !c.$1, orElse: () => (true, '')).$2;
    return (filled, checks.length, filled == checks.length ? null : hint);
  }

  @override
  Widget build(BuildContext context) {
    final (filled, total, hint) = _compute();
    final pct = filled / total;
    if (pct == 1.0) return const SizedBox.shrink(); // 100% — hide the bar

    final color = pct < 0.5
        ? AppColors.accent
        : pct < 0.8
            ? AppColors.secondary
            : AppColors.primary;

    return Container(
      margin: const EdgeInsets.fromLTRB(
          AppSpacing.md, AppSpacing.md, AppSpacing.md, 0),
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
              Text(
                'Профиль толықтығы',
                style: GoogleFonts.nunito(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: AppColors.textSecondary,
                ),
              ),
              const Spacer(),
              Text(
                '${(pct * 100).round()}%',
                style: GoogleFonts.nunito(
                  fontSize: 13,
                  fontWeight: FontWeight.bold,
                  color: color,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          ClipRRect(
            borderRadius: BorderRadius.circular(4),
            child: LinearProgressIndicator(
              value: pct,
              backgroundColor: AppColors.divider,
              // ignore: prefer_const_constructors — color is a runtime variable
              valueColor: AlwaysStoppedAnimation<Color>(color),
              minHeight: 6,
            ),
          ),
          if (hint != null && hint.isNotEmpty) ...[
            const SizedBox(height: 6),
            Text(
              'Кеңес: "$hint" қосыңыз — сәйкестіктер артады',
              style: GoogleFonts.nunito(
                fontSize: 11,
                color: AppColors.textHint,
              ),
            ),
          ],
        ],
      ),
    );
  }
}
