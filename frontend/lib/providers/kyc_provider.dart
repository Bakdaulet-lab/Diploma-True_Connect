import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../core/network/dio_error_message.dart';
import 'auth_provider.dart';

// Returns: 'pending' | 'approved' | 'rejected' | 'banned' | 'not_submitted'
final kycStatusProvider = FutureProvider<String>((ref) async {
  final dio = ref.watch(dioClientProvider).dio;
  try {
    final resp = await dio.get('/kyc/status');
    final data = resp.data is Map ? resp.data['data'] ?? resp.data : resp.data;
    if (data is Map) {
      final status = data['status'] as String?;
      if (status != null && status.isNotEmpty) return status;
    }
    if (data is String && data.isNotEmpty) return data;
    return 'not_submitted';
  } on DioException catch (e) {
    final code = e.response?.statusCode;
    if (code == 404) return 'not_submitted';
    throw dioErrorMessage(e);
  }
});
