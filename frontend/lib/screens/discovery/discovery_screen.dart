import 'package:flutter/material.dart';
import 'package:flutter_card_swiper/flutter_card_swiper.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:haptic_feedback/haptic_feedback.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../widgets/halal_pattern_painter.dart';
import '../../widgets/niyyah_badge.dart';
import '../../widgets/trust_score_badge.dart';
import '../profile/profile_detail_screen.dart';

class DiscoveryScreen extends StatefulWidget {
  const DiscoveryScreen({super.key});

  @override
  State<DiscoveryScreen> createState() => _DiscoveryScreenState();
}

class _DiscoveryScreenState extends State<DiscoveryScreen> {
  final _controller = CardSwiperController();

  // Demo candidates — replace with real API data
  final List<Map<String, dynamic>> _candidates = [
    {
      'id': '1',
      'name': 'Айгерім',
      'age': 24,
      'city': 'Алматы',
      'trustScore': 87,
      'niyyah': 'nikah_year',
      'madhab': 'hanafi',
      'isKycVerified': true,
      'imageUrl':
          'https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=800',
      'bio':
          'Отбасын бағалаймын. Өнерді жақсы көремін. Жақсы адамды іздеймін.',
    },
    {
      'id': '2',
      'name': 'Назерке',
      'age': 26,
      'city': 'Астана',
      'trustScore': 73,
      'niyyah': 'serious_marriage',
      'madhab': 'hanafi',
      'isKycVerified': true,
      'imageUrl':
          'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=800',
      'bio': 'Жеке кәсіпкер. Саяхатты ұнатамын. Жауапты адаммен танысқым келеді.',
    },
    {
      'id': '3',
      'name': 'Дана',
      'age': 23,
      'city': 'Шымкент',
      'trustScore': 55,
      'niyyah': 'friendship',
      'madhab': 'shafi',
      'isKycVerified': false,
      'imageUrl':
          'https://images.unsplash.com/photo-1531746020798-e6953c6e8e04?w=800',
      'bio': 'Дәрігер болып жұмыс жасаймын. Кітапты жақсы көремін.',
    },
  ];

  bool _onSwipe(
    int previousIndex,
    int? currentIndex,
    CardSwiperDirection direction,
  ) {
    if (direction == CardSwiperDirection.right) {
      Haptics.vibrate(HapticsType.success);
    } else if (direction == CardSwiperDirection.left) {
      Haptics.vibrate(HapticsType.medium);
    }
    return true;
  }

  void _onCardTap(int index) async {
    Haptics.vibrate(HapticsType.selection);
    final result = await Navigator.push<String>(
      context,
      PageRouteBuilder(
        transitionDuration: const Duration(milliseconds: 400),
        pageBuilder: (_, animation, __) => FadeTransition(
          opacity: animation,
          child: ProfileDetailScreen(profile: _candidates[index]),
        ),
      ),
    );
    if (!mounted) return;
    if (result == 'like') _controller.swipeRight();
    if (result == 'pass') _controller.swipeLeft();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: _buildAppBar(),
      body: SafeArea(
        child: Column(
          children: [
            Expanded(
              child: _candidates.isEmpty
                  ? _buildEmpty()
                  : CardSwiper(
                      controller: _controller,
                      cardsCount: _candidates.length,
                      onSwipe: _onSwipe,
                      isLoop: false,
                      numberOfCardsDisplayed:
                          _candidates.length > 2 ? 3 : _candidates.length,
                      backCardOffset: const Offset(0, 32),
                      padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
                      cardBuilder: (_, index, hOff, vOff) =>
                          _ProfileCard(
                            candidate: _candidates[index],
                            onTap: () => _onCardTap(index),
                          ),
                    ),
            ),

            // Action buttons — rectangular, not round (NOT Tinder)
            _buildActionRow(),
          ],
        ),
      ),
    );
  }

  PreferredSizeWidget _buildAppBar() {
    return AppBar(
      backgroundColor: AppColors.primaryDark,
      centerTitle: true,
      bottom: const PreferredSize(
        preferredSize: Size.fromHeight(1),
        child: Divider(
          height: 1,
          thickness: 1,
          color: AppColors.goldBorder,
        ),
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
          onPressed: () {},
          tooltip: 'Сүзгі',
        ),
      ],
    );
  }

  Widget _buildActionRow() {
    return Padding(
      padding: const EdgeInsets.fromLTRB(
          AppSpacing.xl, AppSpacing.sm, AppSpacing.xl, AppSpacing.lg),
      child: Row(
        children: [
          // Pass — outlined, NOT red
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
          // Like — primary filled
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
              fontSize: 14,
              color: AppColors.textSecondary,
            ),
          ),
        ],
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
    final score = candidate['trustScore'] as int;
    final niyyah =
        NiyyahTypeExt.fromString(candidate['niyyah'] as String?);
    final madhab = candidate['madhab'] as String? ?? '';
    final isKyc = candidate['isKycVerified'] as bool? ?? false;

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
            // Photo — top 58%
            Expanded(
              flex: 58,
              child: _CardPhoto(
                imageUrl: candidate['imageUrl'] as String,
                score: score,
                isKyc: isKyc,
              ),
            ),

            // Info panel — bottom 42%
            Expanded(
              flex: 42,
              child: _CardInfo(
                name: candidate['name'] as String,
                age: candidate['age'] as int,
                city: candidate['city'] as String,
                bio: candidate['bio'] as String? ?? '',
                niyyah: niyyah,
                madhab: madhab,
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

  const _CardPhoto({
    required this.imageUrl,
    required this.score,
    required this.isKyc,
  });

  @override
  Widget build(BuildContext context) {
    return Stack(
      fit: StackFit.expand,
      children: [
        // Photo
        Image.network(
          imageUrl,
          fit: BoxFit.cover,
          errorBuilder: (_, __, ___) => Container(
            color: AppColors.surfaceVariant,
            child: const Icon(
              Icons.person,
              size: 80,
              color: AppColors.textHint,
            ),
          ),
        ),

        // Bottom gradient overlay
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

        // Trust Score badge — top right
        Positioned(
          top: 14,
          right: 14,
          child: AnimatedTrustScoreBadge(score: score, size: 48),
        ),

        // KYC badge — bottom left on photo
        if (isKyc)
          Positioned(
            bottom: 12,
            left: 14,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
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

  const _CardInfo({
    required this.name,
    required this.age,
    required this.city,
    required this.bio,
    required this.niyyah,
    required this.madhab,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 14, 16, 14),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Name + age
          Row(
            children: [
              Expanded(
                child: Text(
                  '$name, $age',
                  style: GoogleFonts.nunito(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),

          const SizedBox(height: 4),

          // City
          Row(
            children: [
              const Icon(Icons.location_on_outlined,
                  size: 14, color: AppColors.textHint),
              const SizedBox(width: 3),
              Text(
                city,
                style: GoogleFonts.nunito(
                  fontSize: 13,
                  color: AppColors.textSecondary,
                ),
              ),
            ],
          ),

          const SizedBox(height: 10),

          // Chips row: Niyyah + Madhab
          Wrap(
            spacing: 6,
            runSpacing: 6,
            children: [
              NiyyahBadge(niyyah: niyyah),
              if (madhab.isNotEmpty) MadhabBadge(madhab: madhab),
            ],
          ),

          const SizedBox(height: 10),

          // Bio — max 2 lines
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
  late Animation<double> _scale;

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 120),
      lowerBound: 0.97,
      upperBound: 1.0,
    )..value = 1.0;
    _scale = _ctrl;
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
        scale: _scale,
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
