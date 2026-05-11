import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../core/constants/api_constants.dart';
import 'auth_provider.dart';

// ─── Swipe / Candidates ───────────────────────────────────────────────────────

class MatchingNotifier
    extends StateNotifier<AsyncValue<List<Map<String, dynamic>>>> {
  final Dio _dio;

  MatchingNotifier(this._dio) : super(const AsyncValue.loading()) {
    load();
  }

  Future<void> load() async {
    state = const AsyncValue.loading();
    try {
      final resp = await _dio.get(ApiConstants.candidates);
      final list = (resp.data as List<dynamic>)
          .map((e) => e as Map<String, dynamic>)
          .toList();
      state = AsyncValue.data(list);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }

  // Returns true when a mutual match is formed.
  Future<bool> like(String userId) => _swipe(userId, 'like');
  Future<bool> pass(String userId) => _swipe(userId, 'pass');

  Future<bool> _swipe(String userId, String action) async {
    try {
      final resp = await _dio.post(ApiConstants.swipe, data: {
        'target_user_id': userId,
        'action': action,
      });
      return resp.data?['matched'] as bool? ?? false;
    } catch (_) {
      return false;
    }
  }
}

final matchingNotifierProvider = StateNotifierProvider<MatchingNotifier,
    AsyncValue<List<Map<String, dynamic>>>>((ref) {
  return MatchingNotifier(ref.watch(dioClientProvider).dio);
});

// ─── Match list ───────────────────────────────────────────────────────────────

final matchesListProvider =
    FutureProvider<List<Map<String, dynamic>>>((ref) async {
  final dio = ref.watch(dioClientProvider).dio;
  final resp = await dio.get(ApiConstants.matches);
  return (resp.data as List<dynamic>)
      .map((e) => e as Map<String, dynamic>)
      .toList();
});
