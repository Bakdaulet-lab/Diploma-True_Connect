import 'package:flutter/material.dart';
import '../core/theme/app_colors.dart';

class TrustScoreBadge extends StatelessWidget {
  final int score;
  final double size;

  const TrustScoreBadge({super.key, required this.score, this.size = 44});

  Color get _color => AppColors.trustColor(score);
  bool get _isPremium => score >= 80;

  @override
  Widget build(BuildContext context) {
    return Tooltip(
      message:
          'Рейтинг доверия основан на верификации KYC и реальных встречах',
      child: Container(
        width: size,
        height: size,
        decoration: BoxDecoration(
          color: _color,
          shape: BoxShape.circle,
          border: _isPremium
              ? Border.all(color: AppColors.secondary, width: 2)
              : null,
          boxShadow: [
            BoxShadow(
              color: _color.withValues(alpha: 0.4),
              blurRadius: 8,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Text(
              '$score',
              style: TextStyle(
                color: Colors.white,
                fontSize: size * 0.32,
                fontWeight: FontWeight.bold,
                height: 1,
              ),
            ),
            Text(
              '★',
              style: TextStyle(
                color: Colors.white.withValues(alpha: 0.9),
                fontSize: size * 0.2,
                height: 1,
              ),
            ),
          ],
        ),
      ),
    );
  }
}

// Animated counter version for first-show
class AnimatedTrustScoreBadge extends StatefulWidget {
  final int score;
  final double size;

  const AnimatedTrustScoreBadge({
    super.key,
    required this.score,
    this.size = 44,
  });

  @override
  State<AnimatedTrustScoreBadge> createState() =>
      _AnimatedTrustScoreBadgeState();
}

class _AnimatedTrustScoreBadgeState extends State<AnimatedTrustScoreBadge>
    with SingleTickerProviderStateMixin {
  late AnimationController _ctrl;
  late Animation<int> _counter;

  @override
  void initState() {
    super.initState();
    _ctrl = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 800),
    );
    _counter = IntTween(begin: 0, end: widget.score).animate(
      CurvedAnimation(parent: _ctrl, curve: Curves.easeOut),
    );
    _ctrl.forward();
  }

  @override
  void dispose() {
    _ctrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _counter,
      builder: (_, __) => TrustScoreBadge(
        score: _counter.value,
        size: widget.size,
      ),
    );
  }
}
