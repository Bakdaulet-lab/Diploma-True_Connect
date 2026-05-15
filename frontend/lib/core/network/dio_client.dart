import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../constants/api_constants.dart';
import '../services/snack_bar_service.dart';

class DioClient {
  late final Dio _dio;
  final FlutterSecureStorage _storage;
  late final Dio _refreshDio;

  DioClient({FlutterSecureStorage? storage})
      : _storage = storage ?? const FlutterSecureStorage() {
    
    _dio = Dio(BaseOptions(
      baseUrl: ApiConstants.baseUrl,
      connectTimeout: ApiConstants.timeout,
      receiveTimeout: ApiConstants.timeout,
      headers: {'Content-Type': 'application/json'},
    ));

    _refreshDio = Dio(BaseOptions(
      baseUrl: ApiConstants.baseUrl,
      connectTimeout: ApiConstants.timeout,
      receiveTimeout: ApiConstants.timeout,
      headers: {'Content-Type': 'application/json'},
    ));

    _dio.interceptors.add(_AuthInterceptor(_storage, _refreshDio, _dio));
    _dio.interceptors.add(_ErrorInterceptor());

    if (kDebugMode) {
      _dio.interceptors.add(LogInterceptor(
        requestBody: true,
        responseBody: true,
        logPrint: (o) => debugPrint('[DioClient] $o'),
      ));
    }
  }

  Dio get dio => _dio;
}

class _AuthInterceptor extends Interceptor {
  final FlutterSecureStorage _storage;
  final Dio _refreshDio;
  final Dio _mainDio;

  _AuthInterceptor(this._storage, this._refreshDio, this._mainDio);

  @override
  Future<void> onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    final token = await _storage.read(key: 'access_token');

    if (token != null) {
      options.headers['Authorization'] = 'Bearer $token';
    }

    handler.next(options);
  }

  @override
  Future<void> onError(
    DioException err,
    ErrorInterceptorHandler handler,
  ) async {
    if (err.response?.statusCode == 401) {
      final refreshToken = await _storage.read(key: 'refresh_token');

      if (refreshToken == null) {
        handler.next(err);
        return;
      }

      try {
        final resp = await _refreshDio.post(
          ApiConstants.authRefresh,
          data: {
            'refresh_token': refreshToken, // 🔥 ВАЖНО
          },
        );

        final data = resp.data as Map<String, dynamic>;
        final authData = data['data'] ?? data;

        final newAccess = authData['access_token'] as String?;
        final newRefresh = authData['refresh_token'] as String?;

        if (newAccess != null) {
          await _storage.write(key: 'access_token', value: newAccess);
        }
        if (newRefresh != null) {
          await _storage.write(key: 'refresh_token', value: newRefresh);
        }

        final opts = err.requestOptions;

        opts.headers['Authorization'] = 'Bearer $newAccess';

        final retryResp = await _mainDio.fetch(opts); // 🔥 FIX

        handler.resolve(retryResp);
      } on DioException catch (e) {
        await _storage.deleteAll();
        handler.next(e);
      }
    } else {
      handler.next(err);
    }
  }
}

class _ErrorInterceptor extends Interceptor {
  @override
  void onError(DioException err, ErrorInterceptorHandler handler) {
    // 401 is handled by _AuthInterceptor; skip it here.
    final status = err.response?.statusCode;
    if (status == 401) {
      handler.next(err);
      return;
    }

    final String message;
    if (err.type == DioExceptionType.connectionError ||
        err.type == DioExceptionType.connectionTimeout ||
        err.type == DioExceptionType.receiveTimeout ||
        err.type == DioExceptionType.sendTimeout) {
      message = 'Интернет байланысы жоқ';
    } else if (status != null && status >= 500) {
      message = 'Сервер қатесі ($status). Кейінірек қайталаңыз';
    } else {
      handler.next(err);
      return;
    }

    showErrorSnackBar(message);
    handler.next(err);
  }
}