import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../widgets/halal_pattern_painter.dart';

class NiyyahSelectionScreen extends StatefulWidget {
  const NiyyahSelectionScreen({super.key});

  @override
  State<NiyyahSelectionScreen> createState() => _NiyyahSelectionScreenState();
}

class _NiyyahSelectionScreenState extends State<NiyyahSelectionScreen> {
  String? _selected;

  void _continue() {
    if (_selected == null) return;
    context.go('/home');
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.primaryDark,
      body: Stack(
        children: [
          // Pattern on dark background
          const Positioned.fill(
            child: CustomPaint(
              painter: HalalPatternPainter(
                color: AppColors.primaryLight,
                opacity: 0.06,
              ),
            ),
          ),

          SafeArea(
            child: Padding(
              padding: const EdgeInsets.symmetric(horizontal: AppSpacing.xl),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  const SizedBox(height: AppSpacing.xxl),

                  // Header
                  Text(
                    'Мақсатыңыз қандай?',
                    style: GoogleFonts.nunito(
                      fontSize: 28,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                      letterSpacing: -0.5,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Шынайы жауап беріңіз — бұл сізге лайықты адамды\nтабуға көмектеседі',
                    style: GoogleFonts.nunito(
                      fontSize: 14,
                      color: AppColors.secondary,
                      height: 1.5,
                    ),
                  ),

                  const SizedBox(height: AppSpacing.xl),

                  // Cards
                  _NiyyahCard(
                    id: 'nikah_year',
                    emoji: '🌙',
                    title: 'Никях',
                    subtitle: '(бір жыл ішінде)',
                    description:
                        'Отбасының батасымен, шынайы ниетпен',
                    gradient: const LinearGradient(
                      colors: [Color(0xFF1A6B5C), Color(0xFF0D4A3F)],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                    isSelected: _selected == 'nikah_year',
                    onTap: () => setState(() => _selected = 'nikah_year'),
                  ),

                  const SizedBox(height: AppSpacing.md),

                  _NiyyahCard(
                    id: 'serious_marriage',
                    emoji: '💍',
                    title: 'Байыпты қарым-қатынас',
                    subtitle: '',
                    description: 'Қарым-қатынасты қадам-қадам дамытамыз',
                    gradient: const LinearGradient(
                      colors: [Color(0xFFC9860A), Color(0xFF8B5E08)],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                    isSelected: _selected == 'serious_marriage',
                    onTap: () => setState(() => _selected = 'serious_marriage'),
                  ),

                  const SizedBox(height: AppSpacing.md),

                  _NiyyahCard(
                    id: 'friendship',
                    emoji: '🤝',
                    title: 'Танысу',
                    subtitle: '',
                    description: 'Қарым-қатынас шеңберімді кеңейтемін',
                    gradient: const LinearGradient(
                      colors: [Color(0xFF4A4A4A), Color(0xFF2E2E2E)],
                      begin: Alignment.topLeft,
                      end: Alignment.bottomRight,
                    ),
                    isSelected: _selected == 'friendship',
                    onTap: () => setState(() => _selected = 'friendship'),
                  ),

                  const Spacer(),

                  // Continue button
                  AnimatedOpacity(
                    opacity: _selected != null ? 1.0 : 0.4,
                    duration: const Duration(milliseconds: 200),
                    child: ElevatedButton(
                      onPressed: _selected != null ? _continue : null,
                      style: ElevatedButton.styleFrom(
                        backgroundColor: AppColors.primary,
                        disabledBackgroundColor:
                            AppColors.primary.withValues(alpha: 0.5),
                        minimumSize:
                            const Size(double.infinity, 52),
                        shape: const RoundedRectangleBorder(
                          borderRadius: AppRadius.button,
                        ),
                      ),
                      child: Text(
                        'Жалғастыру',
                        style: GoogleFonts.nunito(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                          color: Colors.white,
                        ),
                      ),
                    ),
                  ),

                  const SizedBox(height: AppSpacing.lg),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _NiyyahCard extends StatelessWidget {
  final String id;
  final String emoji;
  final String title;
  final String subtitle;
  final String description;
  final Gradient gradient;
  final bool isSelected;
  final VoidCallback onTap;

  const _NiyyahCard({
    required this.id,
    required this.emoji,
    required this.title,
    required this.subtitle,
    required this.description,
    required this.gradient,
    required this.isSelected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: AnimatedScale(
        scale: isSelected ? 1.02 : 1.0,
        duration: const Duration(milliseconds: 200),
        curve: Curves.easeOut,
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          decoration: BoxDecoration(
            gradient: gradient,
            borderRadius: AppRadius.card,
            border: isSelected
                ? Border.all(color: AppColors.secondary, width: 2)
                : null,
            boxShadow: isSelected
                ? [
                    BoxShadow(
                      color: AppColors.secondary.withValues(alpha: 0.35),
                      blurRadius: 16,
                      offset: const Offset(0, 4),
                    )
                  ]
                : AppShadows.card,
          ),
          child: Stack(
            children: [
              Padding(
                padding: const EdgeInsets.all(AppSpacing.lg),
                child: Row(
                  children: [
                    Text(
                      emoji,
                      style: const TextStyle(fontSize: 48),
                    ),
                    const SizedBox(width: AppSpacing.md),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
  children: [
    Flexible( // <-- Обернули title в Flexible
      child: Text(
        title,
        style: GoogleFonts.nunito(
          fontSize: 18,
          fontWeight: FontWeight.bold,
          color: Colors.white,
        ),
        overflow: TextOverflow.ellipsis, // <-- Добавили троеточие, если текст слишком длинный
      ),
    ),
    if (subtitle.isNotEmpty) ...[
      const SizedBox(width: 6),
      Flexible( // <-- Обернули subtitle в Flexible
        child: Text(
          subtitle,
          style: GoogleFonts.nunito(
            fontSize: 12,
            color: Colors.white70,
          ),
          overflow: TextOverflow.ellipsis, // <-- Добавили троеточие
        ),
      ),
    ],
  ],
),
                          const SizedBox(height: 4),
                          Text(
                            description,
                            style: GoogleFonts.nunito(
                              fontSize: 13,
                              color: Colors.white70,
                              height: 1.4,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),

              // Gold right border accent
              Positioned(
                right: 0,
                top: 0,
                bottom: 0,
                child: Container(
                  width: 4,
                  decoration: BoxDecoration(
                    color: isSelected
                        ? AppColors.secondary
                        : AppColors.secondary.withValues(alpha: 0.4),
                    borderRadius: const BorderRadius.only(
                      topRight: Radius.circular(20),
                      bottomRight: Radius.circular(12),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
