class UserSettings {
  final String userId;
  final bool pushNotifications;
  final bool showOnlineStatus;
  final String distanceUnit;
  final int maxDistanceKm;
  final int ageRangeMin;
  final int ageRangeMax;
  final DateTime? updatedAt;

  const UserSettings({
    this.userId = '',
    this.pushNotifications = true,
    this.showOnlineStatus = true,
    this.distanceUnit = 'km',
    this.maxDistanceKm = 50,
    this.ageRangeMin = 18,
    this.ageRangeMax = 60,
    this.updatedAt,
  });

  factory UserSettings.fromJson(Map<String, dynamic> json) => UserSettings(
        userId: json['user_id'] ?? '',
        pushNotifications: json['push_notifications'] ?? true,
        showOnlineStatus: json['show_online_status'] ?? true,
        distanceUnit: json['distance_unit'] ?? 'km',
        maxDistanceKm: json['max_distance_km'] ?? 50,
        ageRangeMin: json['age_range_min'] ?? 18,
        ageRangeMax: json['age_range_max'] ?? 60,
        updatedAt: json['updated_at'] != null
            ? DateTime.parse(json['updated_at'])
            : null,
      );

  Map<String, dynamic> toJson() => {
        'push_notifications': pushNotifications,
        'show_online_status': showOnlineStatus,
        'distance_unit': distanceUnit,
        'max_distance_km': maxDistanceKm,
        'age_range_min': ageRangeMin,
        'age_range_max': ageRangeMax,
      };

  UserSettings copyWith({
    bool? pushNotifications,
    bool? showOnlineStatus,
    String? distanceUnit,
    int? maxDistanceKm,
    int? ageRangeMin,
    int? ageRangeMax,
  }) =>
      UserSettings(
        userId: userId,
        pushNotifications: pushNotifications ?? this.pushNotifications,
        showOnlineStatus: showOnlineStatus ?? this.showOnlineStatus,
        distanceUnit: distanceUnit ?? this.distanceUnit,
        maxDistanceKm: maxDistanceKm ?? this.maxDistanceKm,
        ageRangeMin: ageRangeMin ?? this.ageRangeMin,
        ageRangeMax: ageRangeMax ?? this.ageRangeMax,
        updatedAt: updatedAt,
      );
}
