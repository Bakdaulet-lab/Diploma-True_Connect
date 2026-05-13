import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import '../core/constants/api_constants.dart';
import '../core/network/dio_client.dart';
import '../core/services/push_service.dart';
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
  }

  // On app start: try to restore tokens from secure storage
  Future<void> _restoreSession() async {
    final token = await _storage.read(key: 'access_token');
    if (token == null) {
      state = const AsyncValue.data(null);
      return;
    }
    // Token exists — attempt to fetch own user from /v1/users/me
    try {
      final resp = await _dio.get('/users/me');
      final data = resp.data as Map<String, dynamic>;
      // Handle wrapped response
      final userData = data['data'] as Map<String, dynamic>? ?? data;
      final user = User.fromJson(userData);
      state = AsyncValue.data(user);
    } catch (_) {
      // Token invalid or server down — stay logged out
      state = const AsyncValue.data(null);
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
      state = AsyncValue.error(_dioMessage(e), st);
    }
  }

  Future<void> logout() async {
    try {
      await _dio.post(ApiConstants.authLogout);
    } catch (_) {
      // Best-effort — always clear local state
    }
    await _storage.deleteAll();
    state = const AsyncValue.data(null);
  }

  Future<void> _saveTokens(Map<String, dynamic> data) async {
    await _persistAuthTokens(data);
    final tokens = AuthTokens.fromJson(data);
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
