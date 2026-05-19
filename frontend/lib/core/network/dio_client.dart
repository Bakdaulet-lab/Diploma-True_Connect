import 'dart:async';
import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../constants/api_constants.dart';
import '../services/app_logger.dart';
import '../services/session_service.dart';
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
    _dio.interceptors.add(_RetryInterceptor(_dio));
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

  // Single-flight: shared across all requests so concurrent 401s trigger
  // exactly one /auth/refresh. Without this, parallel refreshes race and the
  // second one fails (one-time-use refresh-token rotation), wrongly logging
  // the user out mid-session. Resolves to the new access token, or null if
  // the refresh genuinely failed.
  static Future<String?>? _refreshCall;

  static const _retriedKey = '_retriedAfterRefresh';

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
    if (err.response?.statusCode != 401) {
      handler.next(err);
      return;
    }

    // Never try to refresh the refresh call itself, and only retry a given
    // request once to avoid an infinite refresh→401→refresh loop.
    if (err.requestOptions.path == ApiConstants.authRefresh ||
        err.requestOptions.extra[_retriedKey] == true) {
      handler.next(err);
      return;
    }

    AppLogger.info('401 on ${err.requestOptions.path}; attempting token refresh');
    final newToken = await _refreshOnce();
    if (newToken == null) {
      // Refresh genuinely failed; _performRefresh already cleared tokens and
      // signalled the auth layer to route to login.
      AppLogger.warn('Token refresh failed; routing to login');
      handler.next(err);
      return;
    }

    try {
      final opts = err.requestOptions;
      opts.headers['Authorization'] = 'Bearer $newToken';
      opts.extra[_retriedKey] = true;
      final retryResp = await _mainDio.fetch(opts);
      handler.resolve(retryResp);
    } on DioException catch (e) {
      handler.next(e);
    }
  }

  // Coalesces concurrent callers onto one in-flight refresh, then clears the
  // shared slot so a later (post-expiry) 401 can refresh again.
  Future<String?> _refreshOnce() {
    return _refreshCall ??=
        _performRefresh().whenComplete(() => _refreshCall = null);
  }

  Future<String?> _performRefresh() async {
    final refreshToken = await _storage.read(key: 'refresh_token');
    if (refreshToken == null) {
      await _onUnrecoverable();
      return null;
    }
    try {
      final resp = await _refreshDio.post(
        ApiConstants.authRefresh,
        data: {'refresh_token': refreshToken},
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
      return newAccess;
    } on DioException {
      // Refresh token expired/rotated/reused. Clear only the two token keys
      // (not deleteAll(), which would wipe unrelated secure-storage entries
      // and races with other in-flight requests) and signal the auth layer.
      await _onUnrecoverable();
      return null;
    }
  }

  Future<void> _onUnrecoverable() async {
    await _storage.delete(key: 'access_token');
    await _storage.delete(key: 'refresh_token');
    notifySessionExpired();
  }

  /// Extracts a named cookie value from the Set-Cookie response headers.
  String? _extractCookieValue(Headers headers, String name) {
    final cookies = headers['set-cookie'];
    if (cookies == null) return null;
    for (final cookie in cookies) {
      final match = RegExp('$name=([^;]+)').firstMatch(cookie);
      if (match != null) return match.group(1);
    }
    return null;
  }
}

class _RetryInterceptor extends Interceptor {
  final Dio _dio;
  static const _maxRetries = 3;
  static const _retryKey = '_retryCount';

  _RetryInterceptor(this._dio);

  bool _isRetryable(DioException err) {
    final status = err.response?.statusCode;
    return err.type == DioExceptionType.connectionError ||
        err.type == DioExceptionType.connectionTimeout ||
        err.type == DioExceptionType.receiveTimeout ||
        err.type == DioExceptionType.sendTimeout ||
        (status != null && status >= 500);
  }

  @override
  Future<void> onError(DioException err, ErrorInterceptorHandler handler) async {
    if (!_isRetryable(err)) {
      handler.next(err);
      return;
    }

    final retryCount = (err.requestOptions.extra[_retryKey] as int?) ?? 0;
    if (retryCount >= _maxRetries) {
      handler.next(err);
      return;
    }

    final delay = Duration(milliseconds: 100 * (1 << retryCount));
    await Future.delayed(delay);

    final opts = err.requestOptions;
    opts.extra[_retryKey] = retryCount + 1;

    try {
      final response = await _dio.fetch(opts);
      handler.resolve(response);
    } on DioException catch (e) {
      handler.next(e);
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