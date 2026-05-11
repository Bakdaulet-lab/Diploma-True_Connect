class Match {
  final String id;
  final String otherUserId;
  final String otherUserName;
  final String? otherUserAvatarUrl;
  final int otherUserTrustScore;
  final String status;
  final DateTime? niyyahTimerEndsAt;
  final bool familyIntroDone;
  final bool imamConfirmed;
  final DateTime createdAt;

  const Match({
    required this.id,
    required this.otherUserId,
    required this.otherUserName,
    this.otherUserAvatarUrl,
    required this.otherUserTrustScore,
    required this.status,
    this.niyyahTimerEndsAt,
    this.familyIntroDone = false,
    this.imamConfirmed = false,
    required this.createdAt,
  });

  factory Match.fromJson(Map<String, dynamic> json) => Match(
        id: json['id'] as String,
        otherUserId: json['other_user_id'] as String,
        otherUserName: json['other_user_name'] as String? ?? '',
        otherUserAvatarUrl: json['other_user_avatar_url'] as String?,
        otherUserTrustScore: json['other_user_trust_score'] as int? ?? 0,
        status: json['status'] as String? ?? 'matched',
        niyyahTimerEndsAt: json['niyyah_timer_ends_at'] != null
            ? DateTime.tryParse(json['niyyah_timer_ends_at'] as String)
            : null,
        familyIntroDone: json['family_intro_done'] as bool? ?? false,
        imamConfirmed: json['imam_confirmed'] as bool? ?? false,
        createdAt: DateTime.parse(
            json['created_at'] as String? ?? DateTime.now().toIso8601String()),
      );

  int? get daysLeftOnTimer {
    if (niyyahTimerEndsAt == null) return null;
    return niyyahTimerEndsAt!.difference(DateTime.now()).inDays;
  }
}
