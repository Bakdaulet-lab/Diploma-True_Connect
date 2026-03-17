class User {
  final String id;
  final String? phone;
  final String? email;
  final String verificationLevel;
  final String trustStatus;
  final int trustScore;
  final bool isActive;
  final DateTime? lastLoginAt;
  final DateTime createdAt;
  final DateTime updatedAt;

  const User({
    required this.id,
    this.phone,
    this.email,
    this.verificationLevel = 'none',
    this.trustStatus = 'normal',
    this.trustScore = 0,
    this.isActive = true,
    this.lastLoginAt,
    required this.createdAt,
    required this.updatedAt,
  });

  factory User.fromJson(Map<String, dynamic> json) => User(
        id: json['id'] ?? '',
        phone: json['phone'],
        email: json['email'],
        verificationLevel: json['verification_level'] ?? 'none',
        trustStatus: json['trust_status'] ?? 'normal',
        trustScore: json['trust_score'] ?? 0,
        isActive: json['is_active'] ?? true,
        lastLoginAt: json['last_login_at'] != null
            ? DateTime.parse(json['last_login_at'])
            : null,
        createdAt: DateTime.parse(
            json['created_at'] ?? DateTime.now().toIso8601String()),
        updatedAt: DateTime.parse(
            json['updated_at'] ?? DateTime.now().toIso8601String()),
      );

  Map<String, dynamic> toJson() => {
        'id': id,
        'phone': phone,
        'email': email,
        'verification_level': verificationLevel,
        'trust_status': trustStatus,
        'trust_score': trustScore,
        'is_active': isActive,
        'last_login_at': lastLoginAt?.toIso8601String(),
        'created_at': createdAt.toIso8601String(),
        'updated_at': updatedAt.toIso8601String(),
      };
}
