import '../core/constants/api_constants.dart';

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

  // Gender fields (male | female | other)
  final String? gender;
  final String? lookingFor;

  // Halal identity fields
  final String? niyyah;       // nikah_year | serious_marriage | friendship
  final String? madhab;       // hanafi | shafii | maliki | hanbali | none
  final List<String> languages;
  final List<Map<String, dynamic>> prompts;
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
    this.gender,
    this.lookingFor,
    this.niyyah,
    this.madhab,
    this.languages = const [],
    this.prompts = const [],
    this.noPhotoMode = false,
    this.maritalStatus = 'single',
  });

  factory Profile.fromJson(Map<String, dynamic> json) {
    final data = json['data'] is Map
      ? Map<String, dynamic>.from(json['data'] as Map)
      : Map<String, dynamic>.from(json);
    final verificationLevel =
        _readString(data, ['verification_level', 'verificationLevel']) ?? 'none';

    return Profile(
      userId: _readString(data, ['user_id', 'id']) ?? '',
      displayName:
          _readString(data, ['display_name', 'displayName', 'name']) ?? '',
      age: _readInt(data, ['age']),
      city: _readString(data, ['city']),
      bio: _readString(data, ['bio']),
      avatarUrl: ApiConstants.fixImageUrl(
          _readString(data, ['avatar_url', 'avatarUrl', 'imageUrl'])),
      avatarBlurred: _readBool(data, ['avatar_blurred', 'avatarBlurred']),
      trustScore: _readInt(data, ['trust_score', 'trustScore']) ?? 0,
      isKycVerified: _readBool(
        data,
        ['is_kyc_verified', 'isKycVerified'],
      ) || verificationLevel != 'none',
      gender: _readString(data, ['gender']),
      lookingFor: _readString(data, ['looking_for', 'lookingFor']),
      niyyah: _readString(data, ['niyyah']),
      madhab: _readString(data, ['madhab']),
      languages: _readStringList(data, ['languages']),
      noPhotoMode: _readBool(data, ['no_photo_mode', 'noPhotoMode']),
      maritalStatus:
          _readString(data, ['marital_status', 'maritalStatus']) ?? 'single',
      prompts: _readMapList(data, ['prompts']),
    );
  }

  Map<String, dynamic> toJson() => {
        'user_id': userId,
        'display_name': displayName,
        if (age != null) 'age': age,
        if (city != null) 'city': city,
        if (bio != null) 'bio': bio,
        if (avatarUrl != null) 'avatar_url': avatarUrl,
        'no_photo_mode': noPhotoMode,
        if (gender != null) 'gender': gender,
        if (lookingFor != null) 'looking_for': lookingFor,
        if (niyyah != null) 'niyyah': niyyah,
        if (madhab != null) 'madhab': madhab,
        'languages': languages,
        if (prompts.isNotEmpty) 'prompts': prompts,
        'marital_status': maritalStatus,
      };

  Profile copyWith({
    String? displayName,
    int? age,
    String? city,
    String? bio,
    String? avatarUrl,
    String? gender,
    String? lookingFor,
    String? niyyah,
    String? madhab,
    List<String>? languages,
    List<Map<String, dynamic>>? prompts,
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
        gender: gender ?? this.gender,
        lookingFor: lookingFor ?? this.lookingFor,
        niyyah: niyyah ?? this.niyyah,
        madhab: madhab ?? this.madhab,
        languages: languages ?? this.languages,
        prompts: prompts ?? this.prompts,
        noPhotoMode: noPhotoMode ?? this.noPhotoMode,
        maritalStatus: maritalStatus,
      );
}

int? _readInt(Map<String, dynamic> data, List<String> keys) {
  for (final key in keys) {
    final value = data[key];
    if (value is int) return value;
    if (value is num) return value.toInt();
  }
  return null;
}

bool _readBool(Map<String, dynamic> data, List<String> keys) {
  for (final key in keys) {
    final value = data[key];
    if (value is bool) return value;
  }
  return false;
}

String? _readString(Map<String, dynamic> data, List<String> keys) {
  for (final key in keys) {
    final value = data[key];
    if (value is String) return value;
  }
  return null;
}

List<String> _readStringList(Map<String, dynamic> data, List<String> keys) {
  for (final key in keys) {
    final value = data[key];
    if (value is List) {
      return value.whereType<String>().toList();
    }
  }
  return <String>[];
}

List<Map<String, dynamic>> _readMapList(Map<String, dynamic> data, List<String> keys) {
  for (final key in keys) {
    final value = data[key];
    if (value is List) {
      return value
          .whereType<Map>()
          .map((item) => Map<String, dynamic>.from(item))
          .toList();
    }
  }
  return <Map<String, dynamic>>[];
}
