import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';
import '../core/theme/app_colors.dart';
import '../core/theme/app_theme.dart';

enum NiyyahType { nikahYear, seriousMarriage, friendship }

extension NiyyahTypeExt on NiyyahType {
  String get label {
    switch (this) {
      case NiyyahType.nikahYear:
        return 'Никях';
      case NiyyahType.seriousMarriage:
        return 'Серьёзно';
      case NiyyahType.friendship:
        return 'Знакомство';
    }
  }

  String get emoji {
    switch (this) {
      case NiyyahType.nikahYear:
        return '🌙';
      case NiyyahType.seriousMarriage:
        return '💍';
      case NiyyahType.friendship:
        return '🤝';
    }
  }

  Color get backgroundColor {
    switch (this) {
      case NiyyahType.nikahYear:
        return AppColors.primary;
      case NiyyahType.seriousMarriage:
        return AppColors.secondary;
      case NiyyahType.friendship:
        return AppColors.surfaceVariant;
    }
  }

  Color get textColor {
    switch (this) {
      case NiyyahType.nikahYear:
        return Colors.white;
      case NiyyahType.seriousMarriage:
        return AppColors.textPrimary;
      case NiyyahType.friendship:
        return AppColors.textSecondary;
    }
  }

  static NiyyahType fromString(String? value) {
    switch (value) {
      case 'nikah_year':
        return NiyyahType.nikahYear;
      case 'serious_marriage':
        return NiyyahType.seriousMarriage;
      default:
        return NiyyahType.friendship;
    }
  }
}

class NiyyahBadge extends StatelessWidget {
  final NiyyahType niyyah;
  final bool large;

  const NiyyahBadge({super.key, required this.niyyah, this.large = false});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.symmetric(
        horizontal: large ? 14 : 10,
        vertical: large ? 6 : 4,
      ),
      decoration: BoxDecoration(
        color: niyyah.backgroundColor,
        borderRadius: AppRadius.chip,
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            niyyah.emoji,
            style: TextStyle(fontSize: large ? 14 : 11),
          ),
          const SizedBox(width: 4),
          Text(
            niyyah.label,
            style: GoogleFonts.nunito(
              fontSize: large ? 13 : 11,
              fontWeight: FontWeight.w600,
              color: niyyah.textColor,
            ),
          ),
        ],
      ),
    );
  }
}

// Madhab chip
class MadhabBadge extends StatelessWidget {
  final String madhab;

  const MadhabBadge({super.key, required this.madhab});

  String get _label {
    switch (madhab.toLowerCase()) {
      case 'hanafi':
        return 'Ханафи';
      case 'shafi':
        return 'Шафии';
      case 'maliki':
        return 'Маликий';
      case 'hanbali':
        return 'Ханбали';
      default:
        return madhab;
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
      decoration: const BoxDecoration(
        color: AppColors.secondaryLight,
        borderRadius: AppRadius.chip,
      ),
      child: Text(
        _label,
        style: GoogleFonts.nunito(
          fontSize: 11,
          fontWeight: FontWeight.w600,
          color: AppColors.textSecondary,
        ),
      ),
    );
  }
}
