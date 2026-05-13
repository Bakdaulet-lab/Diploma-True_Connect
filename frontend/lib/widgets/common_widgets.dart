import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../core/theme/app_colors.dart';
import '../core/theme/app_theme.dart';
import 'halal_pattern_painter.dart';

class LoadingWidget extends StatelessWidget {
  final String? message;
  const LoadingWidget({super.key, this.message});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const CircularProgressIndicator(color: AppColors.primary),
          if (message != null) ...[
            const SizedBox(height: 16),
            Text(
              message!,
              style: GoogleFonts.nunito(
                fontSize: 14,
                color: AppColors.textSecondary,
              ),
            ),
          ],
        ],
      ),
    );
  }
}

class ErrorRetryWidget extends StatelessWidget {
  final String message;
  final VoidCallback onRetry;
  final IconData icon;

  const ErrorRetryWidget({
    super.key,
    required this.message,
    required this.onRetry,
    this.icon = Icons.wifi_off_outlined,
  });

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 48, color: AppColors.textHint),
            const SizedBox(height: 16),
            Text(
              'Қате орын алды',
              style: GoogleFonts.nunito(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              message,
              textAlign: TextAlign.center,
              style: GoogleFonts.nunito(
                fontSize: 13,
                color: AppColors.textSecondary,
                height: 1.5,
              ),
            ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: onRetry,
              icon: const Icon(Icons.refresh, size: 16),
              label: const Text('Қайтадан көру'),
            ),
          ],
        ),
      ),
    );
  }
}

class SectionDivider extends StatelessWidget {
  final String? label;
  const SectionDivider({super.key, this.label});

  @override
  Widget build(BuildContext context) {
    if (label == null) {
      return const KazakhDivider();
    }
    return Row(
      children: [
        const Expanded(child: KazakhDivider()),
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 12),
          child: Text(
            label!,
            style: GoogleFonts.nunito(
              fontSize: 12,
              color: AppColors.textHint,
              fontWeight: FontWeight.w600,
            ),
          ),
        ),
        const Expanded(child: KazakhDivider()),
      ],
    );
  }
}

class IslamicLoadingOverlay extends StatelessWidget {
  final bool visible;
  final Widget child;

  const IslamicLoadingOverlay({
    super.key,
    required this.visible,
    required this.child,
  });

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: [
        child,
        if (visible)
          Container(
            color: Colors.black.withValues(alpha: 0.4),
            child: const Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  IslamicStarWidget(size: 48, color: AppColors.secondary),
                  SizedBox(height: 16),
                  CircularProgressIndicator(color: AppColors.secondary),
                ],
              ),
            ),
          ),
      ],
    );
  }
}
