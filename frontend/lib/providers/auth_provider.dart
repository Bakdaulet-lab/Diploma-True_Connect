import 'dart:convert';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../core/constants/api_constants.dart';
import '../core/network/dio_client.dart';
import '../core/services/app_logger.dart';
import '../core/services/push_service.dart';
import '../core/services/session_service.dart';
import '../models/user.dart';

// ─── DioClient provider ───────────────────────────────────────────────────────

final dioClientProvider = Provider<DioClient>((ref) => DioClient());

// ─── Auth state ───────────────────────────────────────────────────────────────

// Exposed as AsyncValue<User?> — null = logged out
final authStateProvider =
    StateNotifierProvider<AuthNotifier, AsyncValue<User?>>((ref) {
  return AuthNotifier(
    dio: ref.watch(dioClientProvider).dio,
    storage: const FlutterSecureStorage(),
  );
});

class AuthNotifier extends StateNotifier<AsyncValue<User?>> {
  final Dio _dio;
  final FlutterSecureStorage _storage;

  AuthNotifier({required Dio dio, required FlutterSecureStorage storage})
      : _dio = dio,
        _storage = storage,
        super(const AsyncValue.loading()) {
    _restoreSession();
    sessionExpiredNotifier.addListener(_onSessionExpired);
  }

  // Fired by the network layer when the refresh token is rejected. Force a
  // logged-out state; the router redirects to /auth/login on User == null.
  void _onSessionExpired() {
    if (mounted && state.value != null) {
      AppLogger.warn('Session expired (refresh rejected); forcing logout');
      state = const AsyncValue.data(null);
    }
  }

  @override
  void dispose() {
    sessionExpiredNotifier.removeListener(_onSessionExpired);
    super.dispose();
  }

  // On app start: restore from cached user JSON immediately (no network call),
  // then silently refresh in background so stale data is corrected.
  Future<void> _restoreSession() async {
    final token = await _storage.read(key: 'access_token');
    if (token == null) {
      state = const AsyncValue.data(null);
      return;
    }

    // Fast path: load cached user data from storage to unblock the UI immediately.
    final cachedJson = await _storage.read(key: 'cached_user');
    if (cachedJson != null) {
      try {
        final user = User.fromJson(
            Map<String, dynamic>.from(jsonDecode(cachedJson) as Map));
        state = AsyncValue.data(user);
        // Background refresh — update stale fields without blocking UI.
        _refreshUserInBackground();
        return;
      } catch (_) {
        // Corrupt cache — fall through to network fetch.
      }
    }

    // Slow path: no cache, must fetch from network.
    try {
      final resp = await _dio.get('/users/me');
      final data = resp.data as Map<String, dynamic>;
      final userData = data['data'] as Map<String, dynamic>? ?? data;
      final user = User.fromJson(userData);
      await _storage.write(key: 'cached_user', value: jsonEncode(userData));
      state = AsyncValue.data(user);
    } catch (_) {
      state = const AsyncValue.data(null);
    }
  }

  Future<void> _refreshUserInBackground() async {
    // Use a plain Dio (no auth interceptor) to avoid the race condition where
    // a 401 + failed token refresh triggers _storage.deleteAll() and wipes the
    // access_token for all other in-flight requests (e.g. POST /posts).
    try {
      final token = await _storage.read(key: 'access_token');
      if (token == null) return;
      final plainDio = Dio(BaseOptions(
        baseUrl: ApiConstants.baseUrl,
        connectTimeout: ApiConstants.timeout,
        receiveTimeout: ApiConstants.timeout,
      ));
      final resp = await plainDio.get(
        '/users/me',
        options: Options(headers: {'Authorization': 'Bearer $token'}),
      );
      final data = resp.data as Map<String, dynamic>;
      final userData = data['data'] as Map<String, dynamic>? ?? data;
      final user = User.fromJson(userData);
      await _storage.write(key: 'cached_user', value: jsonEncode(userData));
      if (mounted) state = AsyncValue.data(user);
    } catch (_) {
      // Silently ignore — cached user stays, no token deletion.
    }
  }

  Future<void> register({
    required String phone,
    required String password,
    required String name,
  }) async {
    state = const AsyncValue.loading();
    try {
      final resp = await _dio.post(ApiConstants.authRegister, data: {
        'phone': phone,
        'password': password,
        'name': name,
      });
      final data = resp.data as Map<String, dynamic>;
      final authData = data['data'] as Map<String, dynamic>? ?? data;

      await _persistAuthTokens(authData);

      // Fetch user data separately after the bearer token is stored.
      final userResp = await _dio.get('/users/me');
      final userData = userResp.data is Map ? userResp.data : {};
      final userDataMap = userData['data'] as Map<String, dynamic>? ?? userData;
      authData['user'] = userDataMap;
      authData['refresh_token'] = 'cookie'; // Set via HTTP-only cookie

      await _saveTokens(authData);
      await _patchFcmToken();
    } on DioException catch (e, st) {
      AppLogger.warn('Registration failed', e);
      state = AsyncValue.error(_dioMessage(e), st);
    }
  }

  Future<void> login({
    required String phone,
    required String password,
  }) async {
    state = const AsyncValue.loading();
    try {
      final resp = await _dio.post(ApiConstants.authLogin, data: {
        'phone': phone,
        'password': password,
      });
      final data = resp.data as Map<String, dynamic>;
      final authData = data['data'] as Map<String, dynamic>? ?? data;

      await _persistAuthTokens(authData);

      // Fetch user data separately after the bearer token is stored.
      final userResp = await _dio.get('/users/me');
      final userData = userResp.data is Map ? userResp.data : {};
      final userDataMap = userData['data'] as Map<String, dynamic>? ?? userData;
      authData['user'] = userDataMap;
      authData['refresh_token'] = 'cookie';

      await _saveTokens(authData);
      await _patchFcmToken();
    } on DioException catch (e, st) {
      AppLogger.warn('Login failed', e);
      state = AsyncValue.error(_dioMessage(e), st);
    }
  }

  Future<void> logout() async {
    try {
      await _dio.post(ApiConstants.authLogout);
    } catch (_) {
      // Best-effort — always clear local state
    }
    await _storage.deleteAll(); // also clears cached_user
    state = const AsyncValue.data(null);
  }

  Future<void> _saveTokens(Map<String, dynamic> data) async {
    await _persistAuthTokens(data);
    final tokens = AuthTokens.fromJson(data);
    // Cache user JSON so the next cold start doesn't need a network call.
    if (data['user'] is Map) {
      await _storage.write(
          key: 'cached_user', value: jsonEncode(data['user']));
    }
    state = AsyncValue.data(tokens.user);
  }

  Future<void> _persistAuthTokens(Map<String, dynamic> data) async {
    final accessToken = data['access_token'] as String;
    final refreshToken = data['refresh_token'] as String? ?? 'cookie';

    await _storage.write(key: 'access_token', value: accessToken);
    await _storage.write(key: 'refresh_token', value: refreshToken);
  }

  Future<void> _patchFcmToken() async {
    final token = PushService.fcmToken;
    if (token == null) return;
    try {
      await _dio.patch(ApiConstants.fcmToken, data: {'fcm_token': token});
    } catch (_) {
      // Best-effort — don't fail login if push token upload fails.
    }
  }

  String _dioMessage(DioException e) {
    final data = e.response?.data;
    if (data is Map) {
      final error = data['error'];
      if (error is Map) {
        final message = error['message'];
        if (message is String && message.isNotEmpty) {
          return message;
        }

        final code = error['code'];
        if (code is String && code.isNotEmpty) {
          return code;
        }
      } else if (error is String && error.isNotEmpty) {
        return error;
      }
    }
    return e.message ?? 'Желі қатесі';
  }
}
