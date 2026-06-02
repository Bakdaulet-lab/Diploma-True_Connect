class User {
  final String id;
  final String phone;
  final String? email;
  final bool isKycVerified;
  final bool isAdmin;
  final DateTime createdAt;

  const User({
    required this.id,
    required this.phone,
    this.email,
    required this.isKycVerified,
    required this.isAdmin,
    required this.createdAt,
  });

  factory User.fromJson(Map<String, dynamic> json) => User(
        id: json['id'] as String,
        phone: json['phone'] as String? ?? '',
        email: json['email'] as String?,
        isKycVerified: json['is_kyc_verified'] as bool? ?? false,
        isAdmin: json['is_admin'] as bool? ?? false,
        createdAt: DateTime.parse(
            json['created_at'] as String? ?? DateTime.now().toIso8601String()),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'phone': phone,
        if (email != null) 'email': email,
        'is_kyc_verified': isKycVerified,
        'is_admin': isAdmin,
        'created_at': createdAt.toIso8601String(),
      };
}

class AuthTokens {
  final String accessToken;
  final String refreshToken;
  final User user;

  const AuthTokens({
    required this.accessToken,
    required this.refreshToken,
    required this.user,
  });

  factory AuthTokens.fromJson(Map<String, dynamic> json) => AuthTokens(
        accessToken: json['access_token'] as String? ?? '',
        // On web the refresh token is an HttpOnly cookie that Dio cannot read,
        // so it is absent from the body. Tolerate that instead of throwing — a
        // non-null cast here used to abort login *after* the access token was
        // already saved, which is why login only "worked" after a page refresh.
        refreshToken: json['refresh_token'] as String? ?? '',
        user: User.fromJson(json['user'] as Map<String, dynamic>),
      );
}
