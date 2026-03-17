import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../core/constants/api_constants.dart';
import '../core/network/api_exception.dart';
import '../core/network/dio_client.dart';
import '../models/models.dart';

final apiServiceProvider = Provider<ApiService>((ref) {
  return ApiService(ref.read(dioProvider));
});

class ApiService {
  final Dio _dio;

  ApiService(this._dio);

  // ─── Auth ────────────────────────────────────────

  Future<AuthResponse> register({
    required String phone,
    required String password,
  }) async {
    final response = await _request(
      () => _dio.post(ApiConstants.register, data: {
        'phone': phone,
        'password': password,
      }),
    );
    return AuthResponse.fromJson(response);
  }

  Future<AuthResponse> login({
    required String phone,
    required String password,
  }) async {
    final response = await _request(
      () => _dio.post(ApiConstants.login, data: {
        'phone': phone,
        'password': password,
      }),
    );
    return AuthResponse.fromJson(response);
  }

  Future<AuthResponse> refreshToken(String refreshToken) async {
    final response = await _request(
      () => _dio.post(ApiConstants.refresh, data: {
        'refresh_token': refreshToken,
      }),
    );
    return AuthResponse.fromJson(response);
  }

  Future<void> logout() async {
    await _request(() => _dio.post(ApiConstants.logout));
  }

  // ─── Users ───────────────────────────────────────

  Future<User> getCurrentUser() async {
    final response = await _request(
      () => _dio.get(ApiConstants.usersMe),
    );
    return User.fromJson(response['data'] ?? response);
  }

  Future<void> deleteAccount() async {
    await _request(() => _dio.delete(ApiConstants.usersMe));
  }

  // ─── Profiles ────────────────────────────────────

  Future<Profile> getProfile(String userId) async {
    final response = await _request(
      () => _dio.get(ApiConstants.profile(userId)),
    );
    return Profile.fromJson(response['data'] ?? response);
  }

  Future<Profile> getMyProfile() async {
    final response = await _request(
      () => _dio.get(ApiConstants.profilesMe),
    );
    return Profile.fromJson(response['data'] ?? response);
  }

  Future<Profile> updateProfile(Profile profile) async {
    final response = await _request(
      () => _dio.put(ApiConstants.profilesMe, data: profile.toJson()),
    );
    return Profile.fromJson(response['data'] ?? response);
  }

  Future<List<ProfilePhoto>> getMyPhotos() async {
    final response = await _request(
      () => _dio.get(ApiConstants.profilesMePhotos),
    );
    final list = response['data'] as List? ?? [];
    return list.map((e) => ProfilePhoto.fromJson(e)).toList();
  }

  Future<ProfilePhoto> uploadPhoto(String filePath) async {
    final formData = FormData.fromMap({
      'file': await MultipartFile.fromFile(filePath),
    });
    final response = await _request(
      () => _dio.post(ApiConstants.profilesMePhotos, data: formData),
    );
    return ProfilePhoto.fromJson(response['data'] ?? response);
  }

  Future<void> deletePhoto(String photoId) async {
    await _request(
      () => _dio.delete(ApiConstants.profilePhoto(photoId)),
    );
  }

  // ─── Matching ────────────────────────────────────

  Future<List<Profile>> getCandidates({int page = 1, int limit = 10}) async {
    final response = await _request(
      () => _dio.get(ApiConstants.matchingCandidates, queryParameters: {
        'page': page,
        'limit': limit,
      }),
    );
    final list = response['data'] as List? ?? [];
    return list.map((e) => Profile.fromJson(e)).toList();
  }

  Future<Map<String, dynamic>> likeUser(String targetId) async {
    return await _request(
      () => _dio.post(ApiConstants.matchingLike, data: {
        'target_id': targetId,
      }),
    );
  }

  Future<void> passUser(String targetId) async {
    await _request(
      () => _dio.post(ApiConstants.matchingPass, data: {
        'target_id': targetId,
      }),
    );
  }

  Future<List<Match>> getMatches({int page = 1, int limit = 20}) async {
    final response = await _request(
      () => _dio.get(ApiConstants.matches, queryParameters: {
        'page': page,
        'limit': limit,
      }),
    );
    final list = response['data'] as List? ?? [];
    return list.map((e) => Match.fromJson(e)).toList();
  }

  // ─── Chat ────────────────────────────────────────

  Future<List<Message>> getMessages(String matchId,
      {int page = 1, int limit = 50}) async {
    final response = await _request(
      () => _dio.get(ApiConstants.matchMessages(matchId), queryParameters: {
        'page': page,
        'limit': limit,
      }),
    );
    final list = response['data'] as List? ?? [];
    return list.map((e) => Message.fromJson(e)).toList();
  }

  // ─── Settings ────────────────────────────────────

  Future<UserSettings> getSettings() async {
    final response = await _request(
      () => _dio.get(ApiConstants.settings),
    );
    return UserSettings.fromJson(response['data'] ?? response);
  }

  Future<UserSettings> updateSettings(Map<String, dynamic> updates) async {
    final response = await _request(
      () => _dio.patch(ApiConstants.settings, data: updates),
    );
    return UserSettings.fromJson(response['data'] ?? response);
  }

  // ─── Interactions & Trust ────────────────────────

  Future<Interaction> submitInteraction({
    required String ratedId,
    required int rating,
    required String context,
    String? comment,
  }) async {
    final response = await _request(
      () => _dio.post(ApiConstants.interactions, data: {
        'rated_id': ratedId,
        'rating': rating,
        'context': context,
        if (comment != null) 'comment': comment,
      }),
    );
    return Interaction.fromJson(response['data'] ?? response);
  }

  Future<void> confirmInteraction(String interactionId) async {
    await _request(
      () => _dio.post(ApiConstants.interactionConfirm(interactionId)),
    );
  }

  Future<TrustScore> getReputation(String userId) async {
    final response = await _request(
      () => _dio.get(ApiConstants.reputation(userId)),
    );
    return TrustScore.fromJson(response['data'] ?? response);
  }

  // ─── Posts ───────────────────────────────────────

  Future<List<Post>> getPosts({int page = 1, int limit = 20}) async {
    final response = await _request(
      () => _dio.get(ApiConstants.posts, queryParameters: {
        'page': page,
        'limit': limit,
      }),
    );
    final list = response['data'] as List? ?? [];
    return list.map((e) => Post.fromJson(e)).toList();
  }

  Future<Post> createPost({
    required String content,
    String? mediaPath,
  }) async {
    dynamic data;
    if (mediaPath != null) {
      data = FormData.fromMap({
        'content': content,
        'media': await MultipartFile.fromFile(mediaPath),
      });
    } else {
      data = {'content': content};
    }
    final response = await _request(
      () => _dio.post(ApiConstants.posts, data: data),
    );
    return Post.fromJson(response['data'] ?? response);
  }

  Future<Post> getPost(String postId) async {
    final response = await _request(
      () => _dio.get(ApiConstants.post(postId)),
    );
    return Post.fromJson(response['data'] ?? response);
  }

  Future<void> deletePost(String postId) async {
    await _request(() => _dio.delete(ApiConstants.post(postId)));
  }

  Future<void> likePost(String postId) async {
    await _request(() => _dio.post(ApiConstants.postLike(postId)));
  }

  Future<void> unlikePost(String postId) async {
    await _request(() => _dio.delete(ApiConstants.postLike(postId)));
  }

  Future<List<PostComment>> getComments(String postId,
      {int page = 1, int limit = 20}) async {
    final response = await _request(
      () => _dio.get(ApiConstants.postComments(postId), queryParameters: {
        'page': page,
        'limit': limit,
      }),
    );
    final list = response['data'] as List? ?? [];
    return list.map((e) => PostComment.fromJson(e)).toList();
  }

  Future<PostComment> addComment(String postId, String content) async {
    final response = await _request(
      () => _dio.post(ApiConstants.postComments(postId), data: {
        'content': content,
      }),
    );
    return PostComment.fromJson(response['data'] ?? response);
  }

  // ─── KYC ─────────────────────────────────────────

  Future<void> submitKyc(String filePath) async {
    final formData = FormData.fromMap({
      'document': await MultipartFile.fromFile(filePath),
    });
    await _request(
      () => _dio.post(ApiConstants.kycSubmit, data: formData),
    );
  }

  Future<KycStatus> getKycStatus() async {
    final response = await _request(
      () => _dio.get(ApiConstants.kycStatus),
    );
    return KycStatus.fromJson(response['data'] ?? response);
  }

  // ─── Reports ─────────────────────────────────────

  Future<void> reportUser({
    required String reportedId,
    required String reason,
    String? description,
  }) async {
    await _request(
      () => _dio.post(ApiConstants.reports, data: {
        'reported_id': reportedId,
        'reason': reason,
        if (description != null) 'description': description,
      }),
    );
  }

  // ─── Request Helper ──────────────────────────────

  Future<Map<String, dynamic>> _request(
    Future<Response> Function() request,
  ) async {
    try {
      final response = await request();
      return response.data is Map<String, dynamic>
          ? response.data as Map<String, dynamic>
          : {'data': response.data};
    } on DioException catch (e) {
      if (e.type == DioExceptionType.connectionTimeout ||
          e.type == DioExceptionType.receiveTimeout) {
        throw ApiException.timeout();
      }
      if (e.type == DioExceptionType.connectionError) {
        throw ApiException.network();
      }
      if (e.response != null) {
        throw ApiException.fromJson(
          e.response!.data is Map<String, dynamic>
              ? e.response!.data
              : {'message': e.response!.data.toString()},
          statusCode: e.response!.statusCode,
        );
      }
      throw ApiException.unknown(e.message);
    } catch (e) {
      if (e is ApiException) rethrow;
      throw ApiException.unknown(e.toString());
    }
  }
}
