import 'package:flutter/material.dart';
import '../theme/app_colors.dart';

final scaffoldMessengerKey = GlobalKey<ScaffoldMessengerState>();

void showErrorSnackBar(String message) {
  scaffoldMessengerKey.currentState
    ?..hideCurrentSnackBar()
    ..showSnackBar(
      SnackBar(
        content: Text(message),
        backgroundColor: AppColors.accent,
        behavior: SnackBarBehavior.floating,
        duration: const Duration(seconds: 4),
      ),
    );
}
