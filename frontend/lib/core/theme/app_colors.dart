import 'package:flutter/material.dart';

abstract final class AppColors {
  // Primary — Степная бирюза
  static const primary = Color(0xFF1A6B5C);
  static const primaryDark = Color(0xFF0D4A3F);
  static const primaryLight = Color(0xFFE8F5F2);

  // Secondary — Закатное золото
  static const secondary = Color(0xFFC9860A);
  static const secondaryLight = Color(0xFFFFF3DC);

  // Accent — Рубин
  static const accent = Color(0xFF8B1A2F);
  static const accentLight = Color(0xFFFDE8EC);

  // Backgrounds
  static const background = Color(0xFFFAF7F2);
  static const surface = Color(0xFFFFFFFF);
  static const surfaceVariant = Color(0xFFF0EBE3);

  // Text
  static const textPrimary = Color(0xFF1A1208);
  static const textSecondary = Color(0xFF6B5E4E);
  static const textHint = Color(0xFFA89880);

  // Trust Score levels
  static const trustLow = Color(0xFFCD5C5C);     // 0-39
  static const trustMedium = Color(0xFFDAA520);   // 40-59
  static const trustGood = Color(0xFF2E8B57);     // 60-79
  static const trustHigh = Color(0xFF1A6B5C);     // 80-100 (+ gold border)

  // Chat bubbles
  static const bubbleMine = Color(0xFF1A6B5C);
  static const bubbleTheirs = Color(0xFFFFFFFF);
  static const bubbleMahramMale = Color(0xFF2C5F8A);
  static const bubbleMahram = Color(0xFFC9860A);

  // Divider / Border
  static const divider = Color(0xFFDDD5C8);
  static const goldBorder = Color(0xFFC9860A);

  static Color trustColor(int score) {
    if (score >= 80) return trustHigh;
    if (score >= 60) return trustGood;
    if (score >= 40) return trustMedium;
    return trustLow;
  }
}
