class Interaction {
  final String id;
  final String raterId;
  final String ratedId;
  final int rating;
  final String context;
  final String? comment;
  final bool isVerified;
  final DateTime createdAt;

  const Interaction({
    required this.id,
    required this.raterId,
    required this.ratedId,
    required this.rating,
    required this.context,
    this.comment,
    this.isVerified = false,
    required this.createdAt,
  });

  factory Interaction.fromJson(Map<String, dynamic> json) => Interaction(
        id: json['id'] ?? '',
        raterId: json['rater_id'] ?? '',
        ratedId: json['rated_id'] ?? '',
        rating: json['rating'] ?? 0,
        context: json['context'] ?? '',
        comment: json['comment'],
        isVerified: json['is_verified'] ?? false,
        createdAt: DateTime.parse(
            json['created_at'] ?? DateTime.now().toIso8601String()),
      );

  Map<String, dynamic> toJson() => {
        'rated_id': ratedId,
        'rating': rating,
        'context': context,
        'comment': comment,
      };
}

class TrustScore {
  final String userId;
  final int score;
  final int ratingCount;

  const TrustScore({
    required this.userId,
    this.score = 0,
    this.ratingCount = 0,
  });

  factory TrustScore.fromJson(Map<String, dynamic> json) => TrustScore(
        userId: json['user_id'] ?? '',
        score: json['score'] ?? 0,
        ratingCount: json['rating_count'] ?? 0,
      );

  String get level {
    if (score >= 80) return 'High';
    if (score >= 50) return 'Medium';
    return 'Low';
  }
}

class KycStatus {
  final String status;
  final DateTime? submittedAt;
  final DateTime? reviewedAt;

  const KycStatus({
    this.status = 'none',
    this.submittedAt,
    this.reviewedAt,
  });

  factory KycStatus.fromJson(Map<String, dynamic> json) => KycStatus(
        status: json['status'] ?? 'none',
        submittedAt: json['submitted_at'] != null
            ? DateTime.parse(json['submitted_at'])
            : null,
        reviewedAt: json['reviewed_at'] != null
            ? DateTime.parse(json['reviewed_at'])
            : null,
      );

  bool get isPending => status == 'pending';
  bool get isApproved => status == 'approved';
  bool get isRejected => status == 'rejected';
  bool get isNone => status == 'none';
}
