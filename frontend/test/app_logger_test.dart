import 'package:flutter_test/flutter_test.dart';
import 'package:trueconnect/core/services/app_logger.dart';

void main() {
  group('AppLogger', () {
    test('records message and level', () {
      AppLogger.warn('refresh failed', 'boom');
      final last = AppLogger.recent.last;
      expect(last.level, LogLevel.warn);
      expect(last.message, 'refresh failed');
      expect(last.error, 'boom');
    });

    test('ring buffer is capped (does not grow unbounded)', () {
      for (var i = 0; i < 500; i++) {
        AppLogger.info('entry $i');
      }
      expect(AppLogger.recent.length, lessThanOrEqualTo(200));
      // Oldest entries evicted; the most recent one is retained.
      expect(AppLogger.recent.last.message, 'entry 499');
    });

    test('recent is an unmodifiable view', () {
      AppLogger.info('x');
      expect(() => AppLogger.recent.clear(), throwsUnsupportedError);
    });
  });
}
