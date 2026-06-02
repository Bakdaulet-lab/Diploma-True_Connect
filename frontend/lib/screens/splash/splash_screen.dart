import 'dart:math' as math;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../core/theme/app_colors.dart';
import '../../providers/auth_provider.dart';
import '../../widgets/halal_pattern_painter.dart';

class SplashScreen extends ConsumerStatefulWidget {
  const SplashScreen({super.key});

  @override
  ConsumerState<SplashScreen> createState() => _SplashScreenState();
}

class _SplashScreenState extends ConsumerState<SplashScreen>
    with SingleTickerProviderStateMixin {
  late AnimationController _ctrl;
  late Animation<double> _starRotation;
  late Animation<double> _starScale;
  late Animation<double> _fadeIn;

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1200),
    );

    _starRotation = Tween<double>(begin: 0, end: 2 * math.pi / 8).animate(
      CurvedAnimation(
        parent: _ctrl,
        curve: const Interval(0.0, 0.4, curve: Curves.easeOut),
      ),
    );

    _starScale = TweenSequence<double>([
      TweenSequenceItem(tween: Tween(begin: 0.0, end: 1.15), weight: 40),
      TweenSequenceItem(tween: Tween(begin: 1.15, end: 1.0), weight: 20),
      TweenSequenceItem(tween: ConstantTween(1.0), weight: 40),
    ]).animate(_ctrl);

    _fadeIn = Tween<double>(begin: 0.0, end: 1.0).animate(
      CurvedAnimation(
        parent: _ctrl,
        curve: const Interval(0.4, 0.8, curve: Curves.easeIn),
      ),
    );

    _ctrl.forward().whenComplete(_navigate);
  }

  void _navigate() {
    if (!mounted) return;
    final auth = ref.read(authStateProvider);
    if (auth.isLoading) {
      Future.delayed(const Duration(milliseconds: 200), _navigate);
      return;
    }
    if (auth.valueOrNull != null) {
      context.go('/home');
    } else {
      context.go('/onboarding');
    }
  }

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.primaryDark,
      body: Stack(
        children: [
          // Background pattern
          const Positioned.fill(
            child: CustomPaint(
              painter: HalalPatternPainter(
                color: AppColors.primaryLight,
                opacity: 0.07,
              ),
            ),
          ),

          Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                // Animated Islamic star
                AnimatedBuilder(
                  animation: _ctrl,
                  builder: (_, __) => Transform.scale(
                    scale: _starScale.value,
                    child: Transform.rotate(
                      angle: _starRotation.value,
                      child: const IslamicStarWidget(
                        size: 80,
                        color: AppColors.secondary,
                      ),
                    ),
                  ),
                ),

                const SizedBox(height: 32),

                // App name
                AnimatedBuilder(
                  animation: _fadeIn,
                  builder: (_, __) => Opacity(
                    opacity: _fadeIn.value,
                    child: Column(
                      children: [
                        Text(
                          'TrueConnect',
                          style: GoogleFonts.nunito(
                            fontSize: 32,
                            fontWeight: FontWeight.bold,
                            color: Colors.white,
                            letterSpacing: -0.5,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          'Mahabbat',
                          style: GoogleFonts.nunito(
                            fontSize: 14,
                            fontWeight: FontWeight.w400,
                            color: AppColors.secondary,
                            letterSpacing: 3,
                          ),
                        ),
                        const SizedBox(height: 48),

                        // Gold divider line
                        Container(
                          width: 200,
                          height: 1,
                          color: AppColors.secondary.withValues(alpha: 0.6),
                        ),
                        const SizedBox(height: 24),

                        // Quran quote — Arabic
                        Text(
                          'وَمِنْ آيَاتِهِ أَنْ خَلَقَ لَكُم مِّنْ أَنفُسِكُمْ أَزْوَاجًا',
                          textAlign: TextAlign.center,
                          textDirection: TextDirection.rtl,
                          style: GoogleFonts.amiri(
                            fontSize: 18,
                            color: Colors.white.withValues(alpha: 0.9),
                            height: 1.8,
                          ),
                        ),
                        const SizedBox(height: 8),
                        // Kazakh translation
                        Text(
                          'Оның аяттарының бірі — өздеріңнен жұп жаратқаны',
                          textAlign: TextAlign.center,
                          style: GoogleFonts.nunito(
                            fontSize: 13,
                            color: AppColors.secondary,
                            fontStyle: FontStyle.italic,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          '(Қуран 30:21)',
                          style: GoogleFonts.nunito(
                            fontSize: 11,
                            color: Colors.white38,
                          ),
                        ),

                        const SizedBox(height: 24),
                        Container(
                          width: 200,
                          height: 1,
                          color: AppColors.secondary.withValues(alpha: 0.6),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),

          // Loading indicator at bottom
          Positioned(
            bottom: 48,
            left: 0,
            right: 0,
            child: AnimatedBuilder(
              animation: _ctrl,
              builder: (_, __) => Opacity(
                opacity: _fadeIn.value,
                child: const _PulsingStarLoader(),
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _PulsingStarLoader extends StatefulWidget {
  const _PulsingStarLoader();

  @override
  State<_PulsingStarLoader> createState() => _PulsingStarLoaderState();
}

class _PulsingStarLoaderState extends State<_PulsingStarLoader>
    with SingleTickerProviderStateMixin {
  late AnimationController _pulse;

  @override
  void initState() {
    super.initState();
    _pulse = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 900),
    )..repeat(reverse: true);
  }

  @override
  void dispose() {
    _pulse.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Center(
      child: AnimatedBuilder(
        animation: _pulse,
        builder: (_, __) => Opacity(
          opacity: 0.4 + _pulse.value * 0.6,
          child: Transform.scale(
            scale: 0.8 + _pulse.value * 0.2,
            child: const IslamicStarWidget(
              size: 20,
              color: AppColors.secondary,
            ),
          ),
        ),
      ),
    );
  }
}
