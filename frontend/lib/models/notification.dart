class AppNotification {
  final String id;
  final String userId;
  final String type;
  final String title;
  final String body;
  final Map<String, dynamic> data;
  final bool isRead;
  final DateTime createdAt;

  const AppNotification({
    required this.id,
    required this.userId,
    required this.type,
    required this.title,
    required this.body,
    required this.data,
    required this.isRead,
    required this.createdAt,
  });

  factory AppNotification.fromJson(Map<String, dynamic> json) {
    final raw = json['data'] is Map
        ? Map<String, dynamic>.from(json['data'] as Map)
        : Map<String, dynamic>.from(json);

    return AppNotification(
      id: _str(raw, ['id']) ?? '',
      userId: _str(raw, ['user_id', 'userId']) ?? '',
      type: _str(raw, ['type']) ?? 'system',
      title: _str(raw, ['title']) ?? '',
      body: _str(raw, ['body', 'message']) ?? '',
      data: raw['data'] is Map
          ? Map<String, dynamic>.from(raw['data'] as Map)
          : <String, dynamic>{},
      isRead:
          raw['is_read'] as bool? ?? raw['isRead'] as bool? ?? false,
      createdAt: _parseDate(raw['created_at'] ?? raw['createdAt']),
    );
  }

  AppNotification copyWith({bool? isRead}) => AppNotification(
        id: id,
        userId: userId,
        type: type,
        title: title,
        body: body,
        data: data,
        isRead: isRead ?? this.isRead,
        createdAt: createdAt,
      );
}

String? _str(Map<String, dynamic> d, List<String> keys) {
  for (final k in keys) {
    final v = d[k];
    if (v is String && v.isNotEmpty) return v;
  }
  return null;
}

DateTime _parseDate(dynamic v) {
  if (v is String) return DateTime.tryParse(v)?.toLocal() ?? DateTime.now();
  return DateTime.now();
}
