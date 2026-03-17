class Profile {
  final String userId;
  final String displayName;
  final String bio;
  final String gender;
  final DateTime? birthDate;
  final String city;
  final double latitude;
  final double longitude;
  final String lookingFor;
  final String? avatarUrl;
  final DateTime? createdAt;
  final DateTime? updatedAt;

  const Profile({
    required this.userId,
    this.displayName = '',
    this.bio = '',
    this.gender = '',
    this.birthDate,
    this.city = '',
    this.latitude = 0.0,
    this.longitude = 0.0,
    this.lookingFor = '',
    this.avatarUrl,
    this.createdAt,
    this.updatedAt,
  });

  int? get age {
    if (birthDate == null) return null;
    final now = DateTime.now();
    int age = now.year - birthDate!.year;
    if (now.month < birthDate!.month ||
        (now.month == birthDate!.month && now.day < birthDate!.day)) {
      age--;
    }
    return age;
  }

  factory Profile.fromJson(Map<String, dynamic> json) => Profile(
        userId: json['user_id'] ?? '',
        displayName: json['display_name'] ?? '',
        bio: json['bio'] ?? '',
        gender: json['gender'] ?? '',
        birthDate: json['birth_date'] != null
            ? DateTime.parse(json['birth_date'])
            : null,
        city: json['city'] ?? '',
        latitude: (json['latitude'] ?? 0.0).toDouble(),
        longitude: (json['longitude'] ?? 0.0).toDouble(),
        lookingFor: json['looking_for'] ?? '',
        avatarUrl: json['avatar_url'],
        createdAt: json['created_at'] != null
            ? DateTime.parse(json['created_at'])
            : null,
        updatedAt: json['updated_at'] != null
            ? DateTime.parse(json['updated_at'])
            : null,
      );

  Map<String, dynamic> toJson() => {
        'display_name': displayName,
        'bio': bio,
        'gender': gender,
        'birth_date': birthDate?.toIso8601String().split('T').first,
        'city': city,
        'latitude': latitude,
        'longitude': longitude,
        'looking_for': lookingFor,
      };

  Profile copyWith({
    String? userId,
    String? displayName,
    String? bio,
    String? gender,
    DateTime? birthDate,
    String? city,
    double? latitude,
    double? longitude,
    String? lookingFor,
    String? avatarUrl,
  }) =>
      Profile(
        userId: userId ?? this.userId,
        displayName: displayName ?? this.displayName,
        bio: bio ?? this.bio,
        gender: gender ?? this.gender,
        birthDate: birthDate ?? this.birthDate,
        city: city ?? this.city,
        latitude: latitude ?? this.latitude,
        longitude: longitude ?? this.longitude,
        lookingFor: lookingFor ?? this.lookingFor,
        avatarUrl: avatarUrl ?? this.avatarUrl,
        createdAt: createdAt,
        updatedAt: updatedAt,
      );
}

class ProfilePhoto {
  final String id;
  final String userId;
  final String url;
  final String mediaType;
  final int sortOrder;
  final bool isVerified;
  final DateTime createdAt;

  const ProfilePhoto({
    required this.id,
    required this.userId,
    required this.url,
    this.mediaType = 'image/jpeg',
    this.sortOrder = 0,
    this.isVerified = false,
    required this.createdAt,
  });

  factory ProfilePhoto.fromJson(Map<String, dynamic> json) => ProfilePhoto(
        id: json['id'] ?? '',
        userId: json['user_id'] ?? '',
        url: json['url'] ?? json['object_key'] ?? '',
        mediaType: json['media_type'] ?? 'image/jpeg',
        sortOrder: json['sort_order'] ?? 0,
        isVerified: json['is_verified'] ?? false,
        createdAt: DateTime.parse(
            json['created_at'] ?? DateTime.now().toIso8601String()),
      );
}
