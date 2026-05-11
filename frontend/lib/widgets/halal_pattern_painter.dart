import 'dart:math' as math;
import 'package:flutter/material.dart';
import '../core/theme/app_colors.dart';

// Қошқар мүйіз (бараньи рога) — repeating S-curve motif at low opacity
class HalalPatternPainter extends CustomPainter {
  final Color color;
  final double opacity;

  const HalalPatternPainter({
    this.color = AppColors.primary,
    this.opacity = 0.06,
  });

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = color.withValues(alpha: opacity)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 1.5
      ..strokeCap = StrokeCap.round;

    const cellSize = 48.0;
    final cols = (size.width / cellSize).ceil() + 1;
    final rows = (size.height / cellSize).ceil() + 1;

    for (int row = 0; row < rows; row++) {
      for (int col = 0; col < cols; col++) {
        final ox = col * cellSize;
        final oy = row * cellSize;
        _drawQoshkarMotif(canvas, paint, Offset(ox, oy), cellSize);
      }
    }
  }

  void _drawQoshkarMotif(
      Canvas canvas, Paint paint, Offset origin, double size) {
    final half = size / 2;
    final quarter = size / 4;

    // First S-curve (vertical)
    final path1 = Path();
    path1.moveTo(origin.dx + quarter, origin.dy);
    path1.cubicTo(
      origin.dx + quarter + half, origin.dy,
      origin.dx + quarter - half, origin.dy + size,
      origin.dx + quarter, origin.dy + size,
    );
    canvas.drawPath(path1, paint);

    // Second S-curve (horizontal, rotated 90°)
    final path2 = Path();
    path2.moveTo(origin.dx, origin.dy + quarter);
    path2.cubicTo(
      origin.dx, origin.dy + quarter + half,
      origin.dx + size, origin.dy + quarter - half,
      origin.dx + size, origin.dy + quarter,
    );
    canvas.drawPath(path2, paint);
  }

  @override
  bool shouldRepaint(HalalPatternPainter oldDelegate) =>
      oldDelegate.color != color || oldDelegate.opacity != opacity;
}

// 8-pointed Islamic star
class IslamicStarPainter extends CustomPainter {
  final Color color;
  final double opacity;

  const IslamicStarPainter({
    this.color = AppColors.secondary,
    this.opacity = 1.0,
  });

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = color.withValues(alpha: opacity)
      ..style = PaintingStyle.fill;

    final center = Offset(size.width / 2, size.height / 2);
    final outerR = size.width / 2;
    final innerR = outerR * 0.4;

    final path = Path();
    const points = 8;
    const step = math.pi / points;

    for (int i = 0; i < points * 2; i++) {
      final angle = i * step - math.pi / 2;
      final r = i.isEven ? outerR : innerR;
      final x = center.dx + r * math.cos(angle);
      final y = center.dy + r * math.sin(angle);
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
  bool shouldRepaint(IslamicStarPainter oldDelegate) =>
      oldDelegate.color != color || oldDelegate.opacity != opacity;
}

class IslamicStarWidget extends StatelessWidget {
  final double size;
  final Color color;
  final double opacity;

  const IslamicStarWidget({
    super.key,
    this.size = 24,
    this.color = AppColors.secondary,
    this.opacity = 1.0,
  });

  @override
  Widget build(BuildContext context) {
    return CustomPaint(
      size: Size(size, size),
      painter: IslamicStarPainter(color: color, opacity: opacity),
    );
  }
}

// Kazakh divider — 1px line with diamond center
class KazakhDivider extends StatelessWidget {
  final double indent;

  const KazakhDivider({super.key, this.indent = 0});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.symmetric(horizontal: indent),
      child: Row(
        children: [
          const Expanded(child: Divider(color: AppColors.divider, height: 1)),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 8),
            child: Transform.rotate(
              angle: math.pi / 4,
              child: Container(
                width: 6,
                height: 6,
                decoration: BoxDecoration(
                  color: AppColors.divider,
                  borderRadius: BorderRadius.circular(1),
                ),
              ),
            ),
          ),
          const Expanded(child: Divider(color: AppColors.divider, height: 1)),
        ],
      ),
    );
  }
}
