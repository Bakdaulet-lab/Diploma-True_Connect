import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../core/constants/api_constants.dart';
import '../models/settings.dart';
import 'auth_provider.dart';

class SettingsNotifier extends StateNotifier<AsyncValue<UserSettings>> {
  final Dio _dio;

  SettingsNotifier(this._dio) : super(const AsyncValue.loading()) {
    _load();
  }

  Future<void> _load() async {
    try {
      final resp = await _dio.get(ApiConstants.settings);
      state = AsyncValue.data(
          UserSettings.fromJson(resp.data as Map<String, dynamic>));
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }

  Future<void> update(UserSettings updated) async {
    try {
      await _dio.patch(ApiConstants.settings, data: updated.toJson());
      state = AsyncValue.data(updated);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }
}

final settingsNotifierProvider =
    StateNotifierProvider<SettingsNotifier, AsyncValue<UserSettings>>((ref) {
  return SettingsNotifier(ref.watch(dioClientProvider).dio);
});
