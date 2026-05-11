class Profile {
  final String userId;
  final String displayName;
  final int? age;
  final String? city;
  final String? bio;
  final String? avatarUrl;
  final bool avatarBlurred;
  final int trustScore;
  final bool isKycVerified;

  // Halal identity fields
  final String? niyyah;       // nikah_year | serious_marriage | friendship
  final String? madhab;       // hanafi | shafii | maliki | hanbali | none
  final List<String> languages;
  final bool noPhotoMode;
  final String maritalStatus; // single | married_via_app | divorced

  const Profile({
    required this.userId,
    required this.displayName,
    this.age,
    this.city,
    this.bio,
    this.avatarUrl,
    this.avatarBlurred = false,
    this.trustScore = 0,
    this.isKycVerified = false,
    this.niyyah,
    this.madhab,
    this.languages = const [],
    this.noPhotoMode = false,
    this.maritalStatus = 'single',
  });

  factory Profile.fromJson(Map<String, dynamic> json) => Profile(
        userId: json['user_id'] as String,
        displayName: json['display_name'] as String? ?? '',
        age: json['age'] as int?,
        city: json['city'] as String?,
        bio: json['bio'] as String?,
        avatarUrl: json['avatar_url'] as String?,
        avatarBlurred: json['avatar_blurred'] as bool? ?? false,
        trustScore: json['trust_score'] as int? ?? 0,
        isKycVerified: json['is_kyc_verified'] as bool? ?? false,
        niyyah: json['niyyah'] as String?,
        madhab: json['madhab'] as String?,
        languages: (json['languages'] as List<dynamic>?)
                ?.map((e) => e as String)
                .toList() ??
            [],
        noPhotoMode: json['no_photo_mode'] as bool? ?? false,
        maritalStatus: json['marital_status'] as String? ?? 'single',
      );

  Map<String, dynamic> toJson() => {
        'user_id': userId,
        'display_name': displayName,
        if (age != null) 'age': age,
        if (city != null) 'city': city,
        if (bio != null) 'bio': bio,
        if (avatarUrl != null) 'avatar_url': avatarUrl,
        'no_photo_mode': noPhotoMode,
        if (niyyah != null) 'niyyah': niyyah,
        if (madhab != null) 'madhab': madhab,
        'languages': languages,
        'marital_status': maritalStatus,
      };

  Profile copyWith({
    String? displayName,
    int? age,
    String? city,
    String? bio,
    String? avatarUrl,
    String? niyyah,
    String? madhab,
    List<String>? languages,
    bool? noPhotoMode,
  }) =>
      Profile(
        userId: userId,
        displayName: displayName ?? this.displayName,
        age: age ?? this.age,
        city: city ?? this.city,
        bio: bio ?? this.bio,
        avatarUrl: avatarUrl ?? this.avatarUrl,
        avatarBlurred: avatarBlurred,
        trustScore: trustScore,
        isKycVerified: isKycVerified,
        niyyah: niyyah ?? this.niyyah,
        madhab: madhab ?? this.madhab,
        languages: languages ?? this.languages,
        noPhotoMode: noPhotoMode ?? this.noPhotoMode,
        maritalStatus: maritalStatus,
      );
}
