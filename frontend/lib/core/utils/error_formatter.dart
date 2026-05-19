import 'package:dio/dio.dart';

/// Maps thrown errors to short, user-facing Kazakh messages so screens never
/// render raw exception/stack text (e.g. `DioException [connection error]...`).
abstract final class ErrorFormatter {
  static String message(Object? error) {
    if (error is DioException) {
      switch (error.type) {
        case DioExceptionType.connectionError:
        case DioExceptionType.connectionTimeout:
        case DioExceptionType.receiveTimeout:
        case DioExceptionType.sendTimeout:
          return 'Интернет байланысы жоқ';
        default:
          break;
      }
      final status = error.response?.statusCode;
      if (status != null && status >= 500) {
        return 'Сервер қатесі. Кейінірек қайталаңыз';
      }
      final serverMsg = _serverMessage(error.response?.data);
      if (serverMsg != null) return serverMsg;
      return 'Сұрауды орындау мүмкін болмады';
    }
    return 'Бірдеңе дұрыс болмады';
  }

  static String? _serverMessage(dynamic data) {
    if (data is Map) {
      final error = data['error'];
      if (error is Map) {
        final m = error['message'];
        if (m is String && m.isNotEmpty) return m;
        final c = error['code'];
        if (c is String && c.isNotEmpty) return c;
      } else if (error is String && error.isNotEmpty) {
        return error;
      }
    }
    return null;
  }
}
