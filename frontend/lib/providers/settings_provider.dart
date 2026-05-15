import 'dart:async';

import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../core/constants/api_constants.dart';
import '../core/network/dio_error_message.dart';
import '../core/services/snack_bar_service.dart';
import '../models/settings.dart';
import 'auth_provider.dart';
import 'profile_provider.dart';

class SettingsNotifier extends StateNotifier<AsyncValue<UserSettings>> {
  final Dio _dio;
  final Ref _ref;
  Timer? _saveTimer;
  int _saveVersion = 0;

  SettingsNotifier(this._dio, this._ref) : super(const AsyncValue.loading()) {
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
    // Capture previous known-good state before optimistic update.
    final previous = state.valueOrNull;
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
        _ref.invalidate(ownProfileProvider);
      } on DioException catch (e) {
        if (currentVersion != _saveVersion) return;
        // Revert to previous snapshot so the UI stays consistent.
        if (previous != null) {
          state = AsyncValue.data(previous);
        }
        showErrorSnackBar(dioErrorMessage(e));
      } catch (_) {
        if (currentVersion != _saveVersion) return;
        if (previous != null) {
          state = AsyncValue.data(previous);
        }
      }
    });
  }
}

final settingsNotifierProvider =
    StateNotifierProvider<SettingsNotifier, AsyncValue<UserSettings>>((ref) {
  // Re-create whenever the logged-in user changes so settings are not shared between accounts.
  ref.watch(authStateProvider.select((s) => s.valueOrNull?.id));
  return SettingsNotifier(ref.watch(dioClientProvider).dio, ref);
});
