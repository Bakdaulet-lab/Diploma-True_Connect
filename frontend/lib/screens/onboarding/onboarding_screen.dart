import 'dart:math' as math;
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../widgets/halal_pattern_painter.dart';

class OnboardingScreen extends StatefulWidget {
  const OnboardingScreen({super.key});

  @override
  State<OnboardingScreen> createState() => _OnboardingScreenState();
}

class _OnboardingScreenState extends State<OnboardingScreen> {
  final _pageController = PageController();
  int _currentPage = 0;

  void _next() {
    if (_currentPage < 2) {
      _pageController.nextPage(
        duration: const Duration(milliseconds: 350),
        curve: Curves.easeInOut,
      );
    } else {
      context.go('/niyyah');
    }
  }

  @override
  void dispose() {
    _pageController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      body: Stack(
        children: [
          // Subtle pattern on background
          const Positioned.fill(
            child: CustomPaint(
              painter: HalalPatternPainter(opacity: 0.05),
            ),
          ),

          Column(
            children: [
              Expanded(
                child: PageView(
                  controller: _pageController,
                  onPageChanged: (i) => setState(() => _currentPage = i),
                  children: const [
                    _OnboardingSlide1(),
                    _OnboardingSlide2(),
                    _OnboardingSlide3(),
                  ],
                ),
              ),

              // Dots + button
              Padding(
                padding: const EdgeInsets.fromLTRB(
                    AppSpacing.xl, 0, AppSpacing.xl, 48),
                child: Column(
                  children: [
                    // Gold dots indicator
                    Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: List.generate(3, (i) {
                        final active = i == _currentPage;
                        return AnimatedContainer(
                          duration: const Duration(milliseconds: 250),
                          margin: const EdgeInsets.symmetric(horizontal: 4),
                          width: active ? 24 : 8,
                          height: 8,
                          decoration: BoxDecoration(
                            color: active
                                ? AppColors.secondary
                                : AppColors.secondary.withValues(alpha: 0.3),
                            borderRadius: BorderRadius.circular(4),
                          ),
                        );
                      }),
                    ),

                    const SizedBox(height: 32),

                    ElevatedButton(
                      onPressed: _next,
                      child: Text(
                        _currentPage < 2 ? 'Алға' : 'Бастау',
                      ),
                    ),

                    if (_currentPage < 2) ...[
                      const SizedBox(height: 12),
                      TextButton(
                        onPressed: () => context.go('/auth/login'),
                        child: Text(
                          'Өткізіп жіберу',
                          style: GoogleFonts.nunito(
                            color: AppColors.textHint,
                            fontSize: 14,
                          ),
                        ),
                      ),
                    ],
                  ],
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

// ─── Slide 1: Знакомства с намерением ────────────────────────────────────────

class _OnboardingSlide1 extends StatelessWidget {
  const _OnboardingSlide1();

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.fromLTRB(32, 80, 32, 24),
      physics: const ClampingScrollPhysics(),
      child: Column(
        children: [
          // Illustration: two silhouettes under an arch
          SizedBox(
            height: 220,
            child: CustomPaint(painter: _CoupleUnderArchPainter()),
          ),

          const SizedBox(height: 32),

          // Quran quote
          const _QuranQuote(
            arabic:
                'وَمِنْ آيَاتِهِ أَنْ خَلَقَ لَكُم مِّنْ أَنفُسِكُمْ أَزْوَاجًا',
            kazakh: 'Оның аяттарының бірі — өздеріңнен жұп жаратқаны',
            surah: 'Қуран 30:21',
          ),

          const SizedBox(height: 28),

          Text(
            'Ниетпен танысу',
            textAlign: TextAlign.center,
            style: GoogleFonts.nunito(
              fontSize: 24,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 10),
          Text(
            'Әр пайдаланушы өз мақсатын алдын ала таңдайды',
            textAlign: TextAlign.center,
            style: GoogleFonts.nunito(
              fontSize: 15,
              color: AppColors.textSecondary,
              height: 1.5,
            ),
          ),
        ],
      ),
    );
  }
}

// ─── Slide 2: Безопасность прежде всего ──────────────────────────────────────

class _OnboardingSlide2 extends StatelessWidget {
  const _OnboardingSlide2();

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.fromLTRB(32, 80, 32, 24),
      physics: const ClampingScrollPhysics(),
      child: Column(
        children: [
          SizedBox(
            height: 220,
            child: CustomPaint(painter: _ShieldStarPainter()),
          ),

          const SizedBox(height: 40),

          Text(
            'Қауіпсіздік және сенім',
            textAlign: TextAlign.center,
            style: GoogleFonts.nunito(
              fontSize: 24,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: 10),
          Text(
            'KYC верификация, Trust Score және Махрам қолдауы — сіздің тыныштығыңыз үшін',
            textAlign: TextAlign.center,
            style: GoogleFonts.nunito(
              fontSize: 15,
              color: AppColors.textSecondary,
              height: 1.5,
            ),
          ),

          const SizedBox(height: 28),

          // Feature chips
          const Wrap(
            alignment: WrapAlignment.center,
            spacing: 8,
            runSpacing: 8,
            children: [
              _FeatureChip(icon: '🛡️', label: 'KYC верификация'),
              _FeatureChip(icon: '⭐', label: 'Trust Score'),
              _FeatureChip(icon: '👨‍👩‍👧', label: 'Махрам қатысуы'),
            ],
          ),
        ],
      ),
    );
  }
}

// ─── Slide 3: Для Казахстана ──────────────────────────────────────────────────

class _OnboardingSlide3 extends StatelessWidget {
  const _OnboardingSlide3();

  @override
  Widget build(BuildContext context) {
    // ЗАМЕНИЛИ Padding НА SingleChildScrollView
    return SingleChildScrollView(
      padding: const EdgeInsets.fromLTRB(32, 80, 32, 24),
      physics: const ClampingScrollPhysics(), // Добавили физику как на других слайдах
      child: Column(
        children: [
          SizedBox(
            height: 220,
            child: CustomPaint(painter: _KazakhMapPainter()),
          ),

          const SizedBox(height: 40),

          Text(
            'Астанада жасалған,\nбүкіл Қазақстан үшін',
            textAlign: TextAlign.center,
            style: GoogleFonts.nunito(
              fontSize: 24,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
              height: 1.3,
            ),
          ),
          const SizedBox(height: 10),
          Text(
            'Қазақша, орысша, арабша — сіздің тіліңіз',
            textAlign: TextAlign.center,
            style: GoogleFonts.nunito(
              fontSize: 15,
              color: AppColors.textSecondary,
            ),
          ),

          const SizedBox(height: 28),

          Wrap(
            alignment: WrapAlignment.center,
            spacing: 8,
            runSpacing: 8,
            children: [
              _FeatureChip(
                icon: '🇰🇿',
                label: 'Қазақша',
                onTap: () => ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('Тіл: Қазақша — жақында қосылады')),
                ),
              ),
              _FeatureChip(
                icon: '🌐',
                label: 'Орысша',
                onTap: () => ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('Язык: Русский — скоро будет добавлен')),
                ),
              ),
              _FeatureChip(
                icon: '🕌',
                label: 'Арабша',
                onTap: () => ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(content: Text('Language: Arabic — coming soon')),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

// ─── Helper widgets ───────────────────────────────────────────────────────────

class _FeatureChip extends StatelessWidget {
  final String icon;
  final String label;
  final VoidCallback? onTap;

  const _FeatureChip({required this.icon, required this.label, this.onTap});

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 8),
        decoration: const BoxDecoration(
          color: AppColors.primaryLight,
          borderRadius: AppRadius.chip,
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(icon, style: const TextStyle(fontSize: 14)),
            const SizedBox(width: 6),
            Text(
              label,
              style: GoogleFonts.nunito(
                fontSize: 13,
                fontWeight: FontWeight.w600,
                color: AppColors.primary,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

class _QuranQuote extends StatelessWidget {
  final String arabic;
  final String kazakh;
  final String surah;

  const _QuranQuote({
    required this.arabic,
    required this.kazakh,
    required this.surah,
  });

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        Text(
          arabic,
          textAlign: TextAlign.center,
          style: GoogleFonts.amiri(
            fontSize: 18,
            color: AppColors.primaryDark,
            height: 1.8,
          ),
        ),
        const SizedBox(height: 4),
        Text(
          kazakh,
          textAlign: TextAlign.center,
          style: GoogleFonts.nunito(
            fontSize: 14,
            fontStyle: FontStyle.italic,
            color: AppColors.textSecondary,
            height: 1.6,
          ),
        ),
        const SizedBox(height: 12),
        Text(
          surah,
          style: GoogleFonts.nunito(
            fontSize: 11,
            color: AppColors.textHint,
          ),
        ),
        const SizedBox(height: 12),
        Container(height: 1, color: AppColors.secondary.withValues(alpha: 0.4)),
      ],
    );
  }
}

// ─── Custom Painters for illustrations ───────────────────────────────────────

class _CoupleUnderArchPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final cx = size.width / 2;
    final cy = size.height / 2;

    final paintArch = Paint()
      ..color = AppColors.primary.withValues(alpha: 0.15)
      ..style = PaintingStyle.fill;

    final paintOutline = Paint()
      ..color = AppColors.primary.withValues(alpha: 0.5)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 2;

    // Arch (mihrab shape)
    final archPath = Path();
    archPath.moveTo(cx - 80, size.height - 20);
    archPath.lineTo(cx - 80, cy + 20);
    archPath.arcToPoint(
      Offset(cx + 80, cy + 20),
      radius: const Radius.circular(80),
      largeArc: false,
    );
    archPath.lineTo(cx + 80, size.height - 20);
    archPath.close();
    canvas.drawPath(archPath, paintArch);
    canvas.drawPath(archPath, paintOutline);

    // Left silhouette (female)
    final femPaint = Paint()
      ..color = AppColors.primary.withValues(alpha: 0.6)
      ..style = PaintingStyle.fill;
    _drawSilhouette(canvas, Offset(cx - 30, size.height - 20), femPaint, true);

    // Right silhouette (male)
    final malePaint = Paint()
      ..color = AppColors.primaryDark.withValues(alpha: 0.7)
      ..style = PaintingStyle.fill;
    _drawSilhouette(
        canvas, Offset(cx + 30, size.height - 20), malePaint, false);

    // Islamic star at top of arch
    _drawSmallStar(canvas, Offset(cx, cy - 40), 18);
  }

  void _drawSilhouette(
      Canvas canvas, Offset base, Paint paint, bool isLeft) {
    final path = Path();
    final x = base.dx;
    final y = base.dy;

    // Body
    path.addOval(Rect.fromCenter(
      center: Offset(x, y - 60),
      width: 20,
      height: 26,
    ));
    path.addRect(Rect.fromLTWH(x - 12, y - 48, 24, 40));

    // Head
    path.addOval(Rect.fromCenter(
      center: Offset(x, y - 80),
      width: 18,
      height: 18,
    ));

    canvas.drawPath(path, paint);
  }

  void _drawSmallStar(Canvas canvas, Offset center, double r) {
    final paint = Paint()
      ..color = AppColors.secondary
      ..style = PaintingStyle.fill;

    final path = Path();
    final innerR = r * 0.4;
    const points = 8;
    const step = math.pi / points;

    for (int i = 0; i < points * 2; i++) {
      final angle = i * step - math.pi / 2;
      final radius = i.isEven ? r : innerR;
      final x = center.dx + radius * math.cos(angle);
      final y = center.dy + radius * math.sin(angle);
      if (i == 0) {
        path.moveTo(x, y);
      } else {
        path.lineTo(x, y);
      }
    }
    path.close();
    canvas.drawPath(path, paint);
  }

  @override
  bool shouldRepaint(_CoupleUnderArchPainter oldDelegate) => false;
}

class _ShieldStarPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final cx = size.width / 2;
    final cy = size.height / 2;

    // Shield
    final shieldPaint = Paint()
      ..color = AppColors.primary.withValues(alpha: 0.12)
      ..style = PaintingStyle.fill;
    final shieldOutline = Paint()
      ..color = AppColors.primary.withValues(alpha: 0.5)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 2.5;

    final shield = Path();
    shield.moveTo(cx, cy - 80);
    shield.lineTo(cx + 65, cy - 40);
    shield.lineTo(cx + 65, cy + 20);
    shield.quadraticBezierTo(cx + 65, cy + 80, cx, cy + 90);
    shield.quadraticBezierTo(cx - 65, cy + 80, cx - 65, cy + 20);
    shield.lineTo(cx - 65, cy - 40);
    shield.close();
    canvas.drawPath(shield, shieldPaint);
    canvas.drawPath(shield, shieldOutline);

    // Star inside shield
    _drawStar(canvas, Offset(cx, cy + 5), 38);
  }

  void _drawStar(Canvas canvas, Offset center, double r) {
    final paint = Paint()
      ..color = AppColors.secondary
      ..style = PaintingStyle.fill;
    final path = Path();
    final innerR = r * 0.4;
    const points = 8;
    const step = math.pi / points;
    for (int i = 0; i < points * 2; i++) {
      final angle = i * step - math.pi / 2;
      final radius = i.isEven ? r : innerR;
      final x = center.dx + radius * math.cos(angle);
      final y = center.dy + radius * math.sin(angle);
      if (i == 0) {
        path.moveTo(x, y);
      } else {
        path.lineTo(x, y);
      }
    }
    path.close();
    canvas.drawPath(path, paint);
  }

  @override
  bool shouldRepaint(_ShieldStarPainter oldDelegate) => false;
}

class _KazakhMapPainter extends CustomPainter {
  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = AppColors.primary.withValues(alpha: 0.15)
      ..style = PaintingStyle.fill;
    final outline = Paint()
      ..color = AppColors.primary.withValues(alpha: 0.4)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 2;

    // Simplified abstract Kazakhstan silhouette
    final path = Path();
    final w = size.width;
    final h = size.height;
    path.moveTo(w * 0.15, h * 0.45);
    path.lineTo(w * 0.10, h * 0.35);
    path.lineTo(w * 0.20, h * 0.25);
    path.lineTo(w * 0.38, h * 0.20);
    path.lineTo(w * 0.55, h * 0.18);
    path.lineTo(w * 0.72, h * 0.22);
    path.lineTo(w * 0.88, h * 0.30);
    path.lineTo(w * 0.90, h * 0.45);
    path.lineTo(w * 0.85, h * 0.60);
    path.lineTo(w * 0.70, h * 0.72);
    path.lineTo(w * 0.55, h * 0.78);
    path.lineTo(w * 0.35, h * 0.75);
    path.lineTo(w * 0.22, h * 0.65);
    path.close();
    canvas.drawPath(path, paint);
    canvas.drawPath(path, outline);

    // City dots
    final dotPaint = Paint()
      ..color = AppColors.secondary
      ..style = PaintingStyle.fill;

    final cities = [
      Offset(w * 0.55, h * 0.35), // Astana (center)
      Offset(w * 0.28, h * 0.55), // Almaty area
      Offset(w * 0.72, h * 0.40), // East
      Offset(w * 0.35, h * 0.30), // North-west
      Offset(w * 0.65, h * 0.60), // South-east
    ];

    for (final city in cities) {
      canvas.drawCircle(city, 5, dotPaint);
      canvas.drawCircle(
        city,
        9,
        Paint()
          ..color = AppColors.secondary.withValues(alpha: 0.25)
          ..style = PaintingStyle.fill,
      );
    }

    // Star over Astana
    _drawStar(canvas, cities[0] - const Offset(0, 22), 12);
  }

  void _drawStar(Canvas canvas, Offset center, double r) {
    final paint = Paint()
      ..color = AppColors.secondary
      ..style = PaintingStyle.fill;
    final path = Path();
    final innerR = r * 0.4;
    const points = 8;
    const step = math.pi / points;
    for (int i = 0; i < points * 2; i++) {
      final angle = i * step - math.pi / 2;
      final radius = i.isEven ? r : innerR;
      final x = center.dx + radius * math.cos(angle);
      final y = center.dy + radius * math.sin(angle);
      if (i == 0) {
        path.moveTo(x, y);
      } else {
        path.lineTo(x, y);
      }
    }
    path.close();
    canvas.drawPath(path, paint);
  }

  @override
  bool shouldRepaint(_KazakhMapPainter oldDelegate) => false;
}
