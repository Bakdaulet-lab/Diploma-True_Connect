import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../core/constants/api_constants.dart';
import '../core/network/dio_error_message.dart';
import '../models/profile.dart';
import 'auth_provider.dart';

final ownProfileProvider = FutureProvider<Profile>((ref) async {
  // Re-fetch whenever the logged-in user changes (handles account switching).
  ref.watch(authStateProvider.select((s) => s.valueOrNull?.id));
  final dio = ref.watch(dioClientProvider).dio;
  try {
    final resp = await dio.get(ApiConstants.profile);
    return Profile.fromJson(resp.data as Map<String, dynamic>);
  } on DioException catch (e) {
    throw dioErrorMessage(e);
  }
});
