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
    this.maxAge = 40,
    this.maxDistanceKm = 50,
    this.showMe = true,
    this.modestyLevel = 0,
    this.niyyahFilter,
    this.madhabFilter,
  });

  factory UserSettings.fromJson(Map<String, dynamic> json) => UserSettings(
        minAge: json['min_age'] as int? ?? 18,
        maxAge: json['max_age'] as int? ?? 40,
        maxDistanceKm: json['max_distance_km'] as int? ?? 50,
        showMe: json['show_me'] as bool? ?? true,
        modestyLevel: json['modesty_level'] as int? ?? 0,
        niyyahFilter: json['niyyah_filter'] as String?,
        madhabFilter: json['madhab_filter'] as String?,
      );

  Map<String, dynamic> toJson() => {
        'min_age': minAge,
        'max_age': maxAge,
        'max_distance_km': maxDistanceKm,
        'show_me': showMe,
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
