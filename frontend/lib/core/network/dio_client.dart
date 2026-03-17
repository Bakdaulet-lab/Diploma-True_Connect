import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../constants/api_constants.dart';
import '../storage/token_storage.dart';

final dioProvider = Provider<Dio>((ref) {
  final dio = Dio(BaseOptions(
    baseUrl: ApiConstants.baseUrl,
    connectTimeout: const Duration(seconds: 15),
    receiveTimeout: const Duration(seconds: 15),
    headers: {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
    },
  ));

  dio.interceptors.add(AuthInterceptor(
    dio: dio,
    tokenStorage: ref.read(tokenStorageProvider),
  ));

  dio.interceptors.add(LogInterceptor(
    requestBody: true,
    responseBody: true,
    logPrint: (obj) => print('[DIO] $obj'),
  ));

  return dio;
});

final tokenStorageProvider = Provider<TokenStorage>((ref) => TokenStorage());

class AuthInterceptor extends Interceptor {
  final Dio dio;
  final TokenStorage tokenStorage;
  bool _isRefreshing = false;

  AuthInterceptor({required this.dio, required this.tokenStorage});

  @override
  void onRequest(
    RequestOptions options,
    RequestInterceptorHandler handler,
  ) async {
    // Skip auth for public endpoints
    final publicPaths = [
      ApiConstants.login,
      ApiConstants.register,
      ApiConstants.refresh,
      ApiConstants.health,
    ];
    if (publicPaths.any((path) => options.path.contains(path))) {
      return handler.next(options);
    }

    final token = await tokenStorage.getAccessToken();
    if (token != null) {
      options.headers['Authorization'] = 'Bearer $token';
    }
    handler.next(options);
  }

  @override
  void onError(DioException err, ErrorInterceptorHandler handler) async {
    if (err.response?.statusCode == 401 && !_isRefreshing) {
      _isRefreshing = true;
      try {
        final refreshToken = await tokenStorage.getRefreshToken();
        if (refreshToken == null) {
          await tokenStorage.clearAll();
          return handler.reject(err);
        }

        final response = await Dio().post(
          '${ApiConstants.baseUrl}${ApiConstants.refresh}',
          data: {'refresh_token': refreshToken},
        );

        if (response.statusCode == 200) {
          final data = response.data['data'];
          await tokenStorage.saveTokens(
            accessToken: data['access_token'],
            refreshToken: data['refresh_token'],
          );

          // Retry the original request
          final opts = err.requestOptions;
          opts.headers['Authorization'] = 'Bearer ${data['access_token']}';
          final retryResponse = await dio.fetch(opts);
          return handler.resolve(retryResponse);
        }
      } catch (_) {
        await tokenStorage.clearAll();
      } finally {
        _isRefreshing = false;
      }
    }
    handler.next(err);
  }
}
