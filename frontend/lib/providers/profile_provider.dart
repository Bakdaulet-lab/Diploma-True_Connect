import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/models.dart';
import '../services/api_service.dart';

class ProfileState {
  final Profile? profile;
  final List<ProfilePhoto> photos;
  final bool isLoading;
  final String? error;

  const ProfileState({
    this.profile,
    this.photos = const [],
    this.isLoading = false,
    this.error,
  });

  ProfileState copyWith({
    Profile? profile,
    List<ProfilePhoto>? photos,
    bool? isLoading,
    String? error,
  }) =>
      ProfileState(
        profile: profile ?? this.profile,
        photos: photos ?? this.photos,
        isLoading: isLoading ?? this.isLoading,
        error: error,
      );
}

class ProfileNotifier extends StateNotifier<ProfileState> {
  final ApiService _apiService;

  ProfileNotifier(this._apiService) : super(const ProfileState());

  Future<void> loadMyProfile() async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final profile = await _apiService.getMyProfile();
      final photos = await _apiService.getMyPhotos();
      state = ProfileState(profile: profile, photos: photos);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> updateProfile(Profile profile) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final updated = await _apiService.updateProfile(profile);
      state = state.copyWith(profile: updated, isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> uploadPhoto(String filePath) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final photo = await _apiService.uploadPhoto(filePath);
      state = state.copyWith(
        photos: [...state.photos, photo],
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> deletePhoto(String photoId) async {
    try {
      await _apiService.deletePhoto(photoId);
      state = state.copyWith(
        photos: state.photos.where((p) => p.id != photoId).toList(),
      );
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }
}

final profileProvider =
    StateNotifierProvider<ProfileNotifier, ProfileState>((ref) {
  return ProfileNotifier(ref.read(apiServiceProvider));
});

// Provider for viewing other user's profile
final userProfileProvider =
    FutureProvider.family<Profile, String>((ref, userId) async {
  return ref.read(apiServiceProvider).getProfile(userId);
});
