class UserSettings {
  final int minAge;
  final int maxAge;
  final int maxDistanceKm;
  final bool showMe;

  // Halal matching settings
  final int modestyLevel;       // 0–3
  final String? niyyahFilter;   // restrict to one niyyah only
  final String? madhabFilter;   // restrict to one madhab only

  const UserSettings({
    this.minAge = 18,
    this.maxAge = 60,
    this.maxDistanceKm = 50,
    this.showMe = true,
    this.modestyLevel = 0,
    this.niyyahFilter,
    this.madhabFilter,
  });

  factory UserSettings.fromJson(Map<String, dynamic> json) {
    final data = json['data'] is Map<String, dynamic>
        ? json['data'] as Map<String, dynamic>
        : json;

    return UserSettings(
      minAge: _readInt(data, ['age_range_min', 'AgeRangeMin', 'min_age'], 18),
      maxAge: _readInt(data, ['age_range_max', 'AgeRangeMax', 'max_age'], 60),
      maxDistanceKm: _readInt(
          data, ['max_distance_km', 'MaxDistanceKm'], 50),
      showMe: _readBool(
          data, ['show_online_status', 'ShowOnlineStatus', 'show_me'], true),
      modestyLevel:
          _readInt(data, ['modesty_level', 'ModestyLevel'], 0),
      niyyahFilter: _readString(data, ['niyyah_filter', 'NiyyahFilter']),
      madhabFilter: _readString(data, ['madhab_filter', 'MadhabFilter']),
    );
  }

  Map<String, dynamic> toJson() => {
        'age_range_min': minAge,
        'age_range_max': maxAge,
        'max_distance_km': maxDistanceKm,
        'show_online_status': showMe,
        'modesty_level': modestyLevel,
        if (niyyahFilter != null) 'niyyah_filter': niyyahFilter,
        if (madhabFilter != null) 'madhab_filter': madhabFilter,
      };

  UserSettings copyWith({
    int? minAge,
    int? maxAge,
    int? maxDistanceKm,
    bool? showMe,
    int? modestyLevel,
    String? niyyahFilter,
    String? madhabFilter,
  }) =>
      UserSettings(
        minAge: minAge ?? this.minAge,
        maxAge: maxAge ?? this.maxAge,
        maxDistanceKm: maxDistanceKm ?? this.maxDistanceKm,
        showMe: showMe ?? this.showMe,
        modestyLevel: modestyLevel ?? this.modestyLevel,
        niyyahFilter: niyyahFilter ?? this.niyyahFilter,
        madhabFilter: madhabFilter ?? this.madhabFilter,
      );
}

int _readInt(Map<String, dynamic> data, List<String> keys, int fallback) {
  for (final key in keys) {
    final value = data[key];
    if (value is int) return value;
    if (value is num) return value.toInt();
  }
  return fallback;
}

bool _readBool(Map<String, dynamic> data, List<String> keys, bool fallback) {
  for (final key in keys) {
    final value = data[key];
    if (value is bool) return value;
  }
  return fallback;
}

String? _readString(Map<String, dynamic> data, List<String> keys) {
  for (final key in keys) {
    final value = data[key];
    if (value is String) return value;
  }
  return null;
}
