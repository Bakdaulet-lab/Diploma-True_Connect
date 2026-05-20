import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../core/constants/api_constants.dart';
import '../core/services/app_logger.dart';
import '../core/utils/error_formatter.dart';
import '../models/trust_breakdown.dart';
import 'auth_provider.dart';

/// Fetches the authenticated user's own trust-score breakdown.
/// autoDispose so it re-fetches each time the screen is opened.
final trustBreakdownProvider =
    FutureProvider.autoDispose<TrustBreakdown>((ref) async {
  // Re-fetch when the logged-in user changes (account switching).
  ref.watch(authStateProvider.select((s) => s.valueOrNull?.id));
  final dio = ref.watch(dioClientProvider).dio;
  try {
    final resp = await dio.get(ApiConstants.reputationBreakdown);
    final body = resp.data is Map && (resp.data as Map)['data'] is Map
        ? Map<String, dynamic>.from((resp.data as Map)['data'] as Map)
        : Map<String, dynamic>.from(resp.data as Map);
    return TrustBreakdown.fromJson(body);
  } on DioException catch (e) {
    AppLogger.warn('Trust breakdown fetch failed', e);
    throw ErrorFormatter.message(e);
  }
});
