class AuthResponse {
  final String accessToken;
  final int expiresIn;
  final String userId;
  final String? refreshToken;

  const AuthResponse({
    required this.accessToken,
    required this.expiresIn,
    required this.userId,
    this.refreshToken,
  });

  factory AuthResponse.fromJson(Map<String, dynamic> json) {
    final data = json['data'] ?? json;
    return AuthResponse(
      accessToken: data['access_token'] ?? '',
      expiresIn: data['expires_in'] ?? 900,
      userId: data['user_id'] ?? '',
      refreshToken: data['refresh_token'],
    );
  }
}

class PaginatedResponse<T> {
  final List<T> items;
  final int total;
  final int page;
  final int pageSize;
  final bool hasMore;

  const PaginatedResponse({
    required this.items,
    this.total = 0,
    this.page = 1,
    this.pageSize = 20,
    this.hasMore = false,
  });
}
