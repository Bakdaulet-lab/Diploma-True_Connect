import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../constants/api_constants.dart';

class DioClient {
  late final Dio _dio;
  final FlutterSecureStorage _storage;

  // Separate Dio for token refresh — avoids infinite interceptor loops
  late final Dio _refreshDio;

  DioClient({FlutterSecureStorage? storage})
      : _storage = storage ?? const FlutterSecureStorage() {
    _dio = Dio(BaseOptions(
      baseUrl: ApiConstants.baseUrl,
      connectTimeout: ApiConstants.timeout,
      receiveTimeout: ApiConstants.timeout,
      headers: {'Content-Type': 'application/json'},
      validateStatus: (code) => code != null && code < 500,
    ));

    _refreshDio = Dio(BaseOptions(
      baseUrl: ApiConstants.baseUrl,
      connectTimeout: ApiConstants.timeout,
      receiveTimeout: ApiConstants.timeout,
      headers: {'Content-Type': 'application/json'},
    ));

    _dio.interceptors.add(_AuthInterceptor(_storage, _refreshDio));

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

  _AuthInterceptor(this._storage, this._refreshDio);

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
      // Attempt silent token refresh
      final refreshToken = await _storage.read(key: 'refresh_token');
      if (refreshToken == null) {
        handler.next(err);
        return;
      }

      try {
        final resp = await _refreshDio.post(ApiConstants.authRefresh);
        final data = resp.data as Map<String, dynamic>;
        final authData = data['data'] as Map<String, dynamic>? ?? data;
        final newAccess = authData['access_token'] as String?;

        if (newAccess != null) {
          await _storage.write(key: 'access_token', value: newAccess);
        }

        // Retry original request with new token
        final opts = err.requestOptions;
        opts.headers['Authorization'] = 'Bearer $newAccess';
        final retryResp = await _refreshDio.fetch(opts);
        handler.resolve(retryResp);
      } on DioException catch (e) {
        // Refresh failed — clear tokens, let caller handle redirect
        await _storage.deleteAll();
        handler.next(e);
      }
    } else {
      handler.next(err);
    }
  }
}
