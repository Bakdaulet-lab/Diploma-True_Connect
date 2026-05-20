/// Read-only DTO for GET /reputation/me/breakdown. Manual fromJson (matches
/// the lightweight model style used elsewhere, no build_runner step).
class TrustBreakdown {
  final int score;
  final String badge;
  final int ratingCount;
  final double smoothedRating; // 0-5
  final double baseScore; // smoothed * 20
  final double kycBonus;
  final int reportCount;
  final double reportPenalty;
  final double rawScore;
  final List<TrustRating> recentRatings;

  const TrustBreakdown({
    required this.score,
    required this.badge,
    required this.ratingCount,
    required this.smoothedRating,
    required this.baseScore,
    required this.kycBonus,
    required this.reportCount,
    required this.reportPenalty,
    required this.rawScore,
    required this.recentRatings,
  });

  factory TrustBreakdown.fromJson(Map<String, dynamic> json) {
    // Accepts either the raw breakdown object or the
    // {breakdown:{...}, ratings:[...]} envelope.
    final b = json['breakdown'] is Map
        ? Map<String, dynamic>.from(json['breakdown'] as Map)
        : json;
    final ratingsRaw = json['ratings'] is List ? json['ratings'] as List : const [];

    return TrustBreakdown(
      score: _int(b['score']),
      badge: b['badge'] as String? ?? '',
      ratingCount: _int(b['rating_count']),
      smoothedRating: _double(b['smoothed_rating']),
      baseScore: _double(b['base_score']),
      kycBonus: _double(b['kyc_bonus']),
      reportCount: _int(b['report_count']),
      reportPenalty: _double(b['report_penalty']),
      rawScore: _double(b['raw_score']),
      recentRatings: ratingsRaw
          .whereType<Map>()
          .map((m) => TrustRating.fromJson(Map<String, dynamic>.from(m)))
          .toList(),
    );
  }
}

class TrustRating {
  final String id;
  final int rating;
  final String context;
  final String comment;

  const TrustRating({
    required this.id,
    required this.rating,
    required this.context,
    required this.comment,
  });

  factory TrustRating.fromJson(Map<String, dynamic> json) => TrustRating(
        id: json['id']?.toString() ?? '',
        rating: _int(json['rating']),
        context: json['context'] as String? ?? '',
        comment: json['comment'] as String? ?? '',
      );
}

int _int(dynamic v) {
  if (v is int) return v;
  if (v is double) return v.round();
  if (v is num) return v.toInt();
  return 0;
}

double _double(dynamic v) {
  if (v is double) return v;
  if (v is int) return v.toDouble();
  if (v is num) return v.toDouble();
  return 0.0;
}
