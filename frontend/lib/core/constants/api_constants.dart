class ApiConstants {
  ApiConstants._();

  static const String baseUrl = 'http://localhost:8080';
  static const String apiVersion = '/v1';
  static const String apiBaseUrl = '$baseUrl$apiVersion';

  // Auth
  static const String register = '$apiVersion/auth/register';
  static const String login = '$apiVersion/auth/login';
  static const String refresh = '$apiVersion/auth/refresh';
  static const String logout = '$apiVersion/auth/logout';
  static const String verifyPhone = '$apiVersion/auth/verify-phone';

  // Users
  static const String usersMe = '$apiVersion/users/me';

  // Profiles
  static const String profilesMe = '$apiVersion/profiles/me';
  static const String profilesMePhotos = '$apiVersion/profiles/me/photos';
  static String profile(String id) => '$apiVersion/profiles/$id';
  static String profilePhoto(String photoId) =>
      '$apiVersion/profiles/me/photos/$photoId';

  // Matching
  static const String matchingCandidates = '$apiVersion/matching/candidates';
  static const String matchingLike = '$apiVersion/matching/like';
  static const String matchingPass = '$apiVersion/matching/pass';
  static const String matches = '$apiVersion/matches';

  // Settings
  static const String settings = '$apiVersion/settings';

  // Interactions & Trust
  static const String interactions = '$apiVersion/interactions';
  static String interactionConfirm(String id) =>
      '$apiVersion/interactions/$id/confirm';
  static String reputation(String userId) =>
      '$apiVersion/users/$userId/reputation';

  // Posts
  static const String posts = '$apiVersion/posts';
  static String post(String id) => '$apiVersion/posts/$id';
  static String postLike(String id) => '$apiVersion/posts/$id/like';
  static String postComments(String id) => '$apiVersion/posts/$id/comments';

  // Chat
  static String matchMessages(String matchId) =>
      '$apiVersion/matches/$matchId/messages';
  static const String ws = 'ws://localhost:8080$apiVersion/ws';

  // KYC
  static const String kycSubmit = '$apiVersion/kyc/submit';
  static const String kycStatus = '$apiVersion/kyc/status';

  // Reports
  static const String reports = '$apiVersion/reports';

  // Health
  static const String health = '$apiVersion/health';
}
