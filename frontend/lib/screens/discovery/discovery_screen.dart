import 'package:flutter/material.dart';
import 'package:flutter_card_swiper/flutter_card_swiper.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:haptic_feedback/haptic_feedback.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../providers/matching_provider.dart';
import '../../widgets/halal_pattern_painter.dart';
import '../../widgets/niyyah_badge.dart';
import '../../widgets/trust_score_badge.dart';

class DiscoveryScreen extends ConsumerStatefulWidget {
  const DiscoveryScreen({super.key});

  @override
  ConsumerState<DiscoveryScreen> createState() => _DiscoveryScreenState();
}

class _DiscoveryScreenState extends ConsumerState<DiscoveryScreen> {
  final _controller = CardSwiperController();
  bool _showMatchBanner = false;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Future<void> _onSwipe(
    int previousIndex,
    int? currentIndex,
    CardSwiperDirection direction,
    List<Map<String, dynamic>> candidates,
  ) async {
    if (previousIndex >= candidates.length) return;
    final candidate = candidates[previousIndex];
    final id = candidate['id'] as String? ?? '';

    if (direction == CardSwiperDirection.right) {
      Haptics.vibrate(HapticsType.success);
      final matched =
          await ref.read(matchingNotifierProvider.notifier).like(id);
      if (matched && mounted) {
        setState(() => _showMatchBanner = true);
        Future.delayed(const Duration(seconds: 3), () {
          if (mounted) setState(() => _showMatchBanner = false);
        });
      }
    } else if (direction == CardSwiperDirection.left) {
      Haptics.vibrate(HapticsType.medium);
      ref.read(matchingNotifierProvider.notifier).pass(id);
    }

    if (mounted) setState(() {});
  }

  void _openFilterSheet() => showModalBottomSheet(
        context: context,
        backgroundColor: AppColors.surface,
        shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
        ),
        builder: (_) => const _FilterSheet(),
      );

  @override
  Widget build(BuildContext context) {
    final asyncCandidates = ref.watch(matchingNotifierProvider);

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: _buildAppBar(),
      body: Stack(
        children: [
          SafeArea(
            child: asyncCandidates.when(
              loading: () => const Center(
                child: CircularProgressIndicator(color: AppColors.primary),
              ),
              error: (e, _) => _buildError(e.toString()),
              data: (candidates) => candidates.isEmpty
                  ? _buildEmpty()
                  : Column(
                      children: [
                        Expanded(
                          child: CardSwiper(
                            controller: _controller,
                            cardsCount: candidates.length,
                            onSwipe: (prev, curr, dir) {
                              _onSwipe(prev, curr, dir, candidates);
                              return true;
                            },
                            isLoop: false,
                            numberOfCardsDisplayed:
                                candidates.length > 2 ? 3 : candidates.length,
                            backCardOffset: const Offset(0, 32),
                            padding:
                                const EdgeInsets.fromLTRB(16, 16, 16, 8),
                            cardBuilder: (_, index, hOff, vOff) =>
                                _ProfileCard(
                              candidate: candidates[index],
                              onTap: () {
                                Haptics.vibrate(HapticsType.selection);
                                context.push(
                                  '/profile/${candidates[index]['id']}',
                                );
                              },
                            ),
                          ),
                        ),
                        _buildActionRow(candidates),
                      ],
                    ),
            ),
          ),

          // Match banner
          if (_showMatchBanner)
            Positioned(
              top: 0,
              left: 0,
              right: 0,
              child: _MatchBanner(
                onDismiss: () => setState(() => _showMatchBanner = false),
              ),
            ),
        ],
      ),
    );
  }

  PreferredSizeWidget _buildAppBar() {
    return AppBar(
      backgroundColor: AppColors.primaryDark,
      centerTitle: true,
      bottom: const PreferredSize(
        preferredSize: Size.fromHeight(1),
        child: Divider(height: 1, thickness: 1, color: AppColors.goldBorder),
      ),
      title: Column(
        children: [
          Text(
            'TrueConnect',
            style: GoogleFonts.nunito(
              fontSize: 20,
              fontWeight: FontWeight.bold,
              color: Colors.white,
              letterSpacing: -0.3,
            ),
          ),
          Text(
            'Mahabbat',
            style: GoogleFonts.nunito(
              fontSize: 10,
              color: AppColors.secondary,
              letterSpacing: 2.5,
            ),
          ),
        ],
      ),
      actions: [
        IconButton(
          icon: const IslamicStarWidget(size: 22, color: AppColors.secondary),
          onPressed: _openFilterSheet,
          tooltip: 'Сүзгі',
        ),
      ],
    );
  }

  Widget _buildActionRow(List<Map<String, dynamic>> candidates) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(
          AppSpacing.xl, AppSpacing.sm, AppSpacing.xl, AppSpacing.lg),
      child: Row(
        children: [
          Expanded(
            child: _ActionButton(
              label: '✕  Өткізу',
              onPressed: () {
                Haptics.vibrate(HapticsType.medium);
                _controller.swipeLeft();
              },
              style: _ActionButtonStyle.outlined,
            ),
          ),
          const SizedBox(width: AppSpacing.md),
          Expanded(
            flex: 2,
            child: _ActionButton(
              label: '♥  Ұнайды',
              onPressed: () {
                Haptics.vibrate(HapticsType.success);
                _controller.swipeRight();
              },
              style: _ActionButtonStyle.primary,
            ),
          ),
        ],
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
            'Жаңа кандидаттар жоқ',
            style: GoogleFonts.nunito(
              fontSize: 18,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 8),
          Text(
            'Кейінірек қайта кіріңіз',
            style: GoogleFonts.nunito(
                fontSize: 14, color: AppColors.textSecondary),
          ),
          const SizedBox(height: 24),
          ElevatedButton(
            onPressed: () =>
                ref.read(matchingNotifierProvider.notifier).load(),
            child: const Text('Жаңарту'),
          ),
        ],
      ),
    );
  }

  Widget _buildError(String message) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.wifi_off, size: 48, color: AppColors.textHint),
            const SizedBox(height: 16),
            Text(
              'Желі қатесі',
              style: GoogleFonts.nunito(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              message,
              style: GoogleFonts.nunito(
                  fontSize: 13, color: AppColors.textSecondary),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 24),
            ElevatedButton(
              onPressed: () =>
                  ref.read(matchingNotifierProvider.notifier).load(),
              child: const Text('Қайтадан көру'),
            ),
          ],
        ),
      ),
    );
  }
}

// ─── Profile Card ─────────────────────────────────────────────────────────────

class _ProfileCard extends StatelessWidget {
  final Map<String, dynamic> candidate;
  final VoidCallback onTap;

  const _ProfileCard({required this.candidate, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final score = (candidate['trust_score'] as num?)?.toInt() ??
        (candidate['trustScore'] as num?)?.toInt() ?? 0;
    final niyyah = NiyyahTypeExt.fromString(
        candidate['niyyah'] as String? ?? candidate['niyyah'] as String?);
    final madhab = candidate['madhab'] as String? ?? '';
    final isKyc = candidate['is_kyc_verified'] as bool? ??
        candidate['isKycVerified'] as bool? ?? false;
    final avatarBlurred = candidate['avatar_blurred'] as bool? ?? false;
    final imageUrl = candidate['avatar_url'] as String? ??
        candidate['imageUrl'] as String? ?? '';

    return GestureDetector(
      onTap: onTap,
      child: Container(
        decoration: const BoxDecoration(
          color: AppColors.surface,
          borderRadius: AppRadius.card,
          boxShadow: AppShadows.elevated,
        ),
        clipBehavior: Clip.antiAlias,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Expanded(
              flex: 58,
              child: _CardPhoto(
                imageUrl: imageUrl,
                score: score,
                isKyc: isKyc,
                avatarBlurred: avatarBlurred,
              ),
            ),
            Expanded(
              flex: 42,
              child: _CardInfo(
                name: candidate['name'] as String? ??
                    candidate['display_name'] as String? ?? '',
                age: (candidate['age'] as num?)?.toInt() ?? 0,
                city: candidate['city'] as String? ?? '',
                bio: candidate['bio'] as String? ?? '',
                niyyah: niyyah,
                madhab: madhab,
                languages: (candidate['languages'] as List<dynamic>?)
                        ?.map((e) => e as String)
                        .toList() ??
                    [],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _CardPhoto extends StatelessWidget {
  final String imageUrl;
  final int score;
  final bool isKyc;
  final bool avatarBlurred;

  const _CardPhoto({
    required this.imageUrl,
    required this.score,
    required this.isKyc,
    required this.avatarBlurred,
  });

  @override
  Widget build(BuildContext context) {
    return Stack(
      fit: StackFit.expand,
      children: [
        if (avatarBlurred || imageUrl.isEmpty)
          // No-photo mode placeholder
          Container(
            color: AppColors.surfaceVariant,
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.lock_outline,
                    size: 48, color: AppColors.primary),
                const SizedBox(height: 12),
                Text(
                  'Өзара лайктан кейін\nфото ашылады',
                  textAlign: TextAlign.center,
                  style: GoogleFonts.nunito(
                    fontSize: 13,
                    color: AppColors.textSecondary,
                    height: 1.5,
                  ),
                ),
              ],
            ),
          )
        else
          Image.network(
            imageUrl,
            fit: BoxFit.cover,
            errorBuilder: (_, __, ___) => Container(
              color: AppColors.surfaceVariant,
              child: const Icon(Icons.person,
                  size: 80, color: AppColors.textHint),
            ),
          ),

        if (!avatarBlurred && imageUrl.isNotEmpty)
          Positioned.fill(
            child: DecoratedBox(
              decoration: BoxDecoration(
                gradient: LinearGradient(
                  begin: Alignment.topCenter,
                  end: Alignment.bottomCenter,
                  stops: const [0.55, 1.0],
                  colors: [
                    Colors.transparent,
                    Colors.black.withValues(alpha: 0.55),
                  ],
                ),
              ),
            ),
          ),

        // Trust Score badge
        Positioned(
          top: 14,
          right: 14,
          child: AnimatedTrustScoreBadge(score: score, size: 48),
        ),

        // KYC badge
        if (isKyc)
          Positioned(
            bottom: 12,
            left: 14,
            child: Container(
              padding:
                  const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
              decoration: BoxDecoration(
                color: AppColors.primary.withValues(alpha: 0.9),
                borderRadius: AppRadius.chip,
              ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(Icons.verified, color: Colors.white, size: 13),
                  const SizedBox(width: 4),
                  Text(
                    'Верификацияланған',
                    style: GoogleFonts.nunito(
                      fontSize: 11,
                      fontWeight: FontWeight.w600,
                      color: Colors.white,
                    ),
                  ),
                ],
              ),
            ),
          ),
      ],
    );
  }
}

class _CardInfo extends StatelessWidget {
  final String name;
  final int age;
  final String city;
  final String bio;
  final NiyyahType niyyah;
  final String madhab;
  final List<String> languages;

  const _CardInfo({
    required this.name,
    required this.age,
    required this.city,
    required this.bio,
    required this.niyyah,
    required this.madhab,
    required this.languages,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 14, 16, 14),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            age > 0 ? '$name, $age' : name,
            style: GoogleFonts.nunito(
              fontSize: 20,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),

          const SizedBox(height: 4),

          Row(
            children: [
              const Icon(Icons.location_on_outlined,
                  size: 14, color: AppColors.textHint),
              const SizedBox(width: 3),
              Text(city,
                  style: GoogleFonts.nunito(
                      fontSize: 13, color: AppColors.textSecondary)),
            ],
          ),

          const SizedBox(height: 8),

          // Niyyah + Madhab + languages
          Wrap(
            spacing: 6,
            runSpacing: 4,
            children: [
              NiyyahBadge(niyyah: niyyah),
              if (madhab.isNotEmpty) MadhabBadge(madhab: madhab),
              ...languages.take(2).map(
                    (l) => Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 8, vertical: 3),
                      decoration: BoxDecoration(
                        color: AppColors.surfaceVariant,
                        borderRadius: AppRadius.chip,
                      ),
                      child: Text(
                        l,
                        style: GoogleFonts.nunito(
                            fontSize: 11, color: AppColors.textSecondary),
                      ),
                    ),
                  ),
            ],
          ),

          const SizedBox(height: 8),

          Expanded(
            child: Text(
              bio,
              style: GoogleFonts.nunito(
                fontSize: 13,
                color: AppColors.textSecondary,
                height: 1.4,
              ),
              maxLines: 2,
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
      ),
    );
  }
}

// ─── Match Banner ─────────────────────────────────────────────────────────────

class _MatchBanner extends StatelessWidget {
  final VoidCallback onDismiss;
  const _MatchBanner({required this.onDismiss});

  @override
  Widget build(BuildContext context) {
    return Material(
      color: AppColors.primary,
      child: SafeArea(
        child: Padding(
          padding: const EdgeInsets.symmetric(
              horizontal: AppSpacing.lg, vertical: AppSpacing.md),
          child: Row(
            children: [
              const IslamicStarWidget(size: 24, color: AppColors.secondary),
              const SizedBox(width: 12),
              Expanded(
                child: Text(
                  'Өзара қызығушылық! 90 күн уақыт бар.',
                  style: GoogleFonts.nunito(
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                    color: Colors.white,
                  ),
                ),
              ),
              IconButton(
                icon: const Icon(Icons.close, color: Colors.white, size: 18),
                onPressed: onDismiss,
              ),
            ],
          ),
        ),
      ),
    );
  }
}

// ─── Filter Sheet ─────────────────────────────────────────────────────────────

class _FilterSheet extends StatefulWidget {
  const _FilterSheet();

  @override
  State<_FilterSheet> createState() => _FilterSheetState();
}

class _FilterSheetState extends State<_FilterSheet> {
  String? _niyyah;
  String? _madhab;

  static const _niyyahs = [
    ('nikah_year', 'Никах 🌙'),
    ('serious_marriage', 'Маңызды'),
    ('friendship', 'Достық'),
  ];

  static const _madhabs = [
    ('hanafi', 'Ханафи'),
    ('shafii', 'Шафии'),
    ('maliki', 'Маликий'),
    ('hanbali', 'Ханбали'),
  ];

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(
          AppSpacing.lg, AppSpacing.lg, AppSpacing.lg, AppSpacing.xl),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Text(
                'Сүзгі',
                style: GoogleFonts.nunito(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
              const Spacer(),
              const KazakhDivider(indent: 0),
            ],
          ),
          const SizedBox(height: AppSpacing.lg),

          Text('Ниет',
              style: GoogleFonts.nunito(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: AppColors.textSecondary)),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            children: _niyyahs.map((pair) {
              final selected = _niyyah == pair.$1;
              return FilterChip(
                label: Text(pair.$2,
                    style: GoogleFonts.nunito(
                      color: selected ? Colors.white : AppColors.textPrimary,
                    )),
                selected: selected,
                onSelected: (_) =>
                    setState(() => _niyyah = selected ? null : pair.$1),
                backgroundColor: AppColors.surfaceVariant,
                selectedColor: AppColors.primary,
                checkmarkColor: Colors.white,
              );
            }).toList(),
          ),

          const SizedBox(height: AppSpacing.md),

          Text('Мазхаб',
              style: GoogleFonts.nunito(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: AppColors.textSecondary)),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            children: _madhabs.map((pair) {
              final selected = _madhab == pair.$1;
              return FilterChip(
                label: Text(pair.$2,
                    style: GoogleFonts.nunito(
                      color: selected ? Colors.white : AppColors.textPrimary,
                    )),
                selected: selected,
                onSelected: (_) =>
                    setState(() => _madhab = selected ? null : pair.$1),
                backgroundColor: AppColors.surfaceVariant,
                selectedColor: AppColors.primary,
                checkmarkColor: Colors.white,
              );
            }).toList(),
          ),

          const SizedBox(height: AppSpacing.xl),

          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Қолдану'),
            ),
          ),
        ],
      ),
    );
  }
}

// ─── Action Button ────────────────────────────────────────────────────────────

enum _ActionButtonStyle { primary, outlined }

class _ActionButton extends StatefulWidget {
  final String label;
  final VoidCallback onPressed;
  final _ActionButtonStyle style;

  const _ActionButton({
    required this.label,
    required this.onPressed,
    required this.style,
  });

  @override
  State<_ActionButton> createState() => _ActionButtonState();
}

class _ActionButtonState extends State<_ActionButton>
    with SingleTickerProviderStateMixin {
  late AnimationController _ctrl;

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 120),
      lowerBound: 0.97,
      upperBound: 1.0,
    )..value = 1.0;
  }

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  void _onTapDown(_) => _ctrl.reverse();
  void _onTapUp(_) {
    _ctrl.forward();
    widget.onPressed();
  }

  void _onTapCancel() => _ctrl.forward();

  @override
  Widget build(BuildContext context) {
    final isPrimary = widget.style == _ActionButtonStyle.primary;
    return GestureDetector(
      onTapDown: _onTapDown,
      onTapUp: _onTapUp,
      onTapCancel: _onTapCancel,
      child: ScaleTransition(
        scale: _ctrl,
        child: Container(
          height: 52,
          decoration: BoxDecoration(
            color: isPrimary ? AppColors.primary : Colors.transparent,
            borderRadius: AppRadius.button,
            border: isPrimary
                ? null
                : Border.all(color: AppColors.surfaceVariant, width: 1.5),
          ),
          alignment: Alignment.center,
          child: Text(
            widget.label,
            style: GoogleFonts.nunito(
              fontSize: 15,
              fontWeight: FontWeight.w600,
              color: isPrimary ? Colors.white : AppColors.textSecondary,
            ),
          ),
        ),
      ),
    );
  }
}
