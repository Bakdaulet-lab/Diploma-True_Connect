import 'package:dio/dio.dart';
import 'package:flutter_test/flutter_test.dart' hide ErrorFormatter;
import 'package:trueconnect/core/utils/error_formatter.dart';

void main() {
  final opts = RequestOptions(path: '/x');

  group('ErrorFormatter', () {
    test('connection errors map to offline message', () {
      final e = DioException(
        requestOptions: opts,
        type: DioExceptionType.connectionError,
      );
      expect(ErrorFormatter.message(e), 'Интернет байланысы жоқ');
    });

    test('5xx maps to server error message', () {
      final e = DioException(
        requestOptions: opts,
        response: Response(requestOptions: opts, statusCode: 503),
        type: DioExceptionType.badResponse,
      );
      expect(ErrorFormatter.message(e), 'Сервер қатесі. Кейінірек қайталаңыз');
    });

    test('server-provided error.message is surfaced', () {
      final e = DioException(
        requestOptions: opts,
        response: Response(
          requestOptions: opts,
          statusCode: 422,
          data: {
            'error': {'message': 'Жас шектеуі дұрыс емес'}
          },
        ),
        type: DioExceptionType.badResponse,
      );
      expect(ErrorFormatter.message(e), 'Жас шектеуі дұрыс емес');
    });

    test('non-Dio errors never leak raw toString', () {
      expect(ErrorFormatter.message(StateError('boom')),
          'Бірдеңе дұрыс болмады');
    });
  });
}
