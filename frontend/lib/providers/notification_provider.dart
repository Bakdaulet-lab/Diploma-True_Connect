import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../core/constants/api_constants.dart';
import '../core/network/dio_error_message.dart';
import '../models/notification.dart';
import 'auth_provider.dart';

// ─── Notifications list ───────────────────────────────────────────────────────

class NotificationNotifier
    extends StateNotifier<AsyncValue<List<AppNotification>>> {
  final Dio _dio;

  NotificationNotifier(this._dio) : super(const AsyncValue.loading()) {
    load();
  }

  Future<void> load() async {
    state = const AsyncValue.loading();
    try {
      final resp = await _dio.get(ApiConstants.notifications);
      final list = _parse(resp.data);
      state = AsyncValue.data(list);
    } on DioException catch (e, st) {
      state = AsyncValue.error(dioErrorMessage(e), st);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }

  Future<void> markRead(String id) async {
    try {
      await _dio.patch('${ApiConstants.notifications}/$id/read');
      final current = state.valueOrNull ?? [];
      state = AsyncValue.data(
        current.map((n) => n.id == id ? n.copyWith(isRead: true) : n).toList(),
      );
    } catch (_) {}
  }

  Future<void> markAllRead() async {
    try {
      await _dio.post('${ApiConstants.notifications}/read-all');
      final current = state.valueOrNull ?? [];
      state = AsyncValue.data(
        current.map((n) => n.copyWith(isRead: true)).toList(),
      );
    } catch (_) {}
  }
}

final notificationsProvider = StateNotifierProvider<NotificationNotifier,
    AsyncValue<List<AppNotification>>>((ref) {
  return NotificationNotifier(ref.watch(dioClientProvider).dio);
});

// ─── Unread count ─────────────────────────────────────────────────────────────

final unreadCountProvider = FutureProvider<int>((ref) async {
  final dio = ref.watch(dioClientProvider).dio;
  try {
    final resp = await dio.get('${ApiConstants.notifications}/unread-count');
    final data = resp.data is Map ? resp.data['data'] ?? resp.data : resp.data;
    if (data is Map) {
      final count = data['count'] ?? data['unread_count'] ?? data['unreadCount'];
      if (count is int) return count;
      if (count is num) return count.toInt();
    }
    if (data is int) return data;
    return 0;
  } catch (_) {
    return 0;
  }
});

// ─── Helpers ──────────────────────────────────────────────────────────────────

List<AppNotification> _parse(dynamic payload) {
  final data = payload is Map ? payload['data'] ?? payload : payload;
  if (data is List) {
    return data
        .whereType<Map>()
        .map((e) =>
            AppNotification.fromJson(Map<String, dynamic>.from(e)))
        .toList();
  }
  return [];
}
