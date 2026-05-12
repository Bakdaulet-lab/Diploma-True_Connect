import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../core/constants/api_constants.dart';
import '../core/network/dio_error_message.dart';
import '../models/settings.dart';
import 'auth_provider.dart';

class SettingsNotifier extends StateNotifier<AsyncValue<UserSettings>> {
  final Dio _dio;
  Timer? _saveTimer;
  int _saveVersion = 0;

  SettingsNotifier(this._dio) : super(const AsyncValue.loading()) {
    _load();
  }

  @override
  void dispose() {
    _saveTimer?.cancel();
    super.dispose();
  }

  Future<void> _load() async {
    try {
      final resp = await _dio.get(ApiConstants.settings);
      state = AsyncValue.data(
          UserSettings.fromJson(resp.data as Map<String, dynamic>));
    } on DioException catch (e, st) {
      state = AsyncValue.error(dioErrorMessage(e), st);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }

  Future<void> update(UserSettings updated) async {
    state = AsyncValue.data(updated);

    _saveVersion += 1;
    final currentVersion = _saveVersion;
    _saveTimer?.cancel();
    _saveTimer = Timer(const Duration(milliseconds: 450), () async {
      try {
        final resp = await _dio.patch(ApiConstants.settings, data: updated.toJson());
        if (currentVersion != _saveVersion) return;
        final data = resp.data as Map<String, dynamic>;
        final payload = data['data'] as Map<String, dynamic>? ?? data;
        state = AsyncValue.data(UserSettings.fromJson(payload));
      } on DioException catch (e, st) {
        if (currentVersion == _saveVersion) {
          state = AsyncValue.error(dioErrorMessage(e), st);
        }
      } catch (e, st) {
        if (currentVersion == _saveVersion) {
          state = AsyncValue.error(e, st);
        }
      }
    });
  }
}

final settingsNotifierProvider =
    StateNotifierProvider<SettingsNotifier, AsyncValue<UserSettings>>((ref) {
  return SettingsNotifier(ref.watch(dioClientProvider).dio);
});
