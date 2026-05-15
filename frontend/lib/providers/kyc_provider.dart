import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../core/network/dio_error_message.dart';
import 'auth_provider.dart';

// Returns:
// none | id_verified
final kycStatusProvider = FutureProvider<String>((ref) async {
  // Re-fetch whenever the logged-in user changes so KYC status isn't leaked
  // between accounts on the same device.
  ref.watch(authStateProvider.select((s) => s.valueOrNull?.id));
  final dio = ref.watch(dioClientProvider).dio;

  try {
    final resp = await dio.get('/kyc/status');

    final body = resp.data;
    final data = body is Map
        ? (body['data'] ?? body)
        : null;

    if (data is Map) {
      final level = data['verification_level'] as String?;

      if (level != null && level.isNotEmpty) {
        return level;
      }
    }

    return 'none';
  } on DioException catch (e) {
    final code = e.response?.statusCode;

    if (code == 404) {
      return 'none';
    }

    throw dioErrorMessage(e);
  }
});