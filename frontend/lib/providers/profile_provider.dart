import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../core/constants/api_constants.dart';
import '../core/network/dio_error_message.dart';
import '../models/profile.dart';
import 'auth_provider.dart';

// Own profile — used globally (profile_screen, edit_profile_screen)
final ownProfileProvider = FutureProvider<Profile?>((ref) async {
  final dio = ref.watch(dioClientProvider).dio;
  try {
    final resp = await dio.get(ApiConstants.profile);
    final data = resp.data is Map
        ? Map<String, dynamic>.from(resp.data as Map)
        : <String, dynamic>{};
    return Profile.fromJson(data);
  } on DioException catch (e) {
    throw dioErrorMessage(e);
  }
});

// Any user's profile — keyed by userId
final profileProvider =
    FutureProvider.family<Profile?, String>((ref, userId) async {
  if (userId.isEmpty) return null;
  final dio = ref.watch(dioClientProvider).dio;
  try {
    final resp = await dio.get('${ApiConstants.profiles}/$userId');
    final data = resp.data is Map
        ? Map<String, dynamic>.from(resp.data as Map)
        : <String, dynamic>{};
    return Profile.fromJson(data);
  } on DioException catch (e) {
    throw dioErrorMessage(e);
  }
});
