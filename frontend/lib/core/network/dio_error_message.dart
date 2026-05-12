import 'package:dio/dio.dart';

String dioErrorMessage(Object error) {
  if (error is! DioException) {
    return error.toString();
  }

  final data = error.response?.data;
  if (data is Map) {
    final apiError = data['error'];
    if (apiError is Map) {
      final message = apiError['message'];
      if (message is String && message.isNotEmpty) {
        return message;
      }

      final code = apiError['code'];
      if (code is String && code.isNotEmpty) {
        return code;
      }
    } else if (apiError is String && apiError.isNotEmpty) {
      return apiError;
    }
  }

  return error.message ?? 'Желі қатесі';
}