import 'dart:collection';
import 'dart:developer' as developer;
import 'package:flutter/foundation.dart';

enum LogLevel { debug, info, warn, error }

class LogEntry {
  final DateTime time;
  final LogLevel level;
  final String message;
  final Object? error;

  LogEntry(this.time, this.level, this.message, this.error);

  @override
  String toString() =>
      '${time.toIso8601String()} [${level.name.toUpperCase()}] $message'
      '${error != null ? ' :: $error' : ''}';
}

/// Lightweight structured logger for the Flutter client. Keeps an in-memory
/// ring buffer (inspectable for a debug screen / bug report) and forwards to
/// `dart:developer`. No paid SaaS dependency; a remote sink (e.g. Sentry)
/// could be attached in [_log] behind a --dart-define when desired.
abstract final class AppLogger {
  static const _maxBuffer = 200;
  static final Queue<LogEntry> _buffer = Queue<LogEntry>();

  static List<LogEntry> get recent => List.unmodifiable(_buffer);

  static void debug(String m) => _log(LogLevel.debug, m);
  static void info(String m) => _log(LogLevel.info, m);
  static void warn(String m, [Object? e]) => _log(LogLevel.warn, m, e);
  static void error(String m, [Object? e, StackTrace? s]) =>
      _log(LogLevel.error, m, e, s);

  static void _log(LogLevel level, String message,
      [Object? error, StackTrace? stack]) {
    final entry = LogEntry(DateTime.now(), level, message, error);
    _buffer.addLast(entry);
    while (_buffer.length > _maxBuffer) {
      _buffer.removeFirst();
    }
    developer.log(
      message,
      name: 'TrueConnect',
      error: error,
      stackTrace: stack,
      level: _devLevel(level),
    );
    if (kDebugMode) {
      debugPrint(entry.toString());
    }
  }

  static int _devLevel(LogLevel l) {
    switch (l) {
      case LogLevel.debug:
        return 500;
      case LogLevel.info:
        return 800;
      case LogLevel.warn:
        return 900;
      case LogLevel.error:
        return 1000;
    }
  }
}
