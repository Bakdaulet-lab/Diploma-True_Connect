import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../core/constants/api_constants.dart';
import '../core/network/dio_error_message.dart';
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
      final list = _extractList(resp.data);
      state = AsyncValue.data(list);
    } on DioException catch (e, st) {
      state = AsyncValue.error(dioErrorMessage(e), st);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }

  // Returns true when a mutual match is formed.
  Future<bool> like(String userId) => _submitDecision(userId, ApiConstants.matchLike);
  Future<bool> pass(String userId) => _submitDecision(userId, ApiConstants.matchPass);

  Future<bool> _submitDecision(String userId, String endpoint) async {
    try {
      final resp = await _dio.post(endpoint, data: {
        'target_id': userId,
      });
      final data = _extractMap(resp.data);
      return data['matched'] as bool? ?? false;
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
  try {
    final resp = await dio.get(ApiConstants.matches);
    return _extractList(resp.data);
  } on DioException catch (e) {
    throw dioErrorMessage(e);
  }
});

List<Map<String, dynamic>> _extractList(dynamic payload) {
  final data = payload is Map ? payload['data'] ?? payload : payload;
  if (data is List) {
    return data
        .whereType<Map>()
        .map((item) => _normalizeMatch(Map<String, dynamic>.from(item)))
        .toList();
  }
  return <Map<String, dynamic>>[];
}

Map<String, dynamic> _extractMap(dynamic payload) {
  final data = payload is Map ? payload['data'] ?? payload : payload;
  if (data is Map) {
    return Map<String, dynamic>.from(data);
  }
  return <String, dynamic>{};
}

Map<String, dynamic> _normalizeMatch(Map<String, dynamic> raw) {
  final rawId = raw['id'] ?? raw['user_id'] ?? raw['userId'];
  final otherUser = _extractMap(raw['other_user']);
  final displayName = _readString(otherUser, ['display_name', 'displayName', 'name']);
  final avatarUrl = _readString(otherUser, ['avatar_url', 'avatarUrl', 'imageUrl']);
  final trustScore = _readInt(otherUser, ['trust_score', 'trustScore']) ?? 0;

  return {
    ...raw,
    'id': rawId?.toString() ?? '',
    'other_user': otherUser,
    'other_user_name': displayName,
    'other_user_avatar_url': avatarUrl != null && avatarUrl.isNotEmpty ? avatarUrl : null,
    'other_user_trust_score': trustScore,
    'other_user_public_key': _readString(otherUser, ['public_key', 'publicKey']),
    'imam_confirmed': raw['imam_confirmed'] ?? raw['imamConfirmed'] ?? false,
    'family_intro_done': raw['family_intro_done'] ?? raw['familyIntroDone'] ?? false,
    'niyyah_timer_ends_at': raw['niyyah_timer_ends_at'] ?? raw['niyyahTimerEndsAt'],
  };
}

String? _readString(Map<String, dynamic> data, List<String> keys) {
  for (final key in keys) {
    final value = data[key];
    if (value is String && value.isNotEmpty) {
      return value;
    }
  }
  return null;
}

int? _readInt(Map<String, dynamic> data, List<String> keys) {
  for (final key in keys) {
    final value = data[key];
    if (value is int) return value;
    if (value is num) return value.toInt();
  }
  return null;
}
