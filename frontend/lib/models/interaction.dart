enum InteractionContext { date, meetup, event }

extension InteractionContextLabel on InteractionContext {
  String get label {
    switch (this) {
      case InteractionContext.date:
        return 'кездесу';
      case InteractionContext.meetup:
        return 'кеш';
      case InteractionContext.event:
        return 'іс-шара';
    }
  }

  String get value {
    switch (this) {
      case InteractionContext.date:
        return 'date';
      case InteractionContext.meetup:
        return 'meetup';
      case InteractionContext.event:
        return 'event';
    }
  }
}

class Interaction {
  final String id;
  final String fromUserId;
  final String toUserId;
  final InteractionContext context;
  final int rating;
  final String? comment;
  final DateTime? confirmedAt;

  const Interaction({
    required this.id,
    required this.fromUserId,
    required this.toUserId,
    required this.context,
    required this.rating,
    this.comment,
    this.confirmedAt,
  });

  factory Interaction.fromJson(Map<String, dynamic> json) {
    final raw = json['data'] is Map
        ? Map<String, dynamic>.from(json['data'])
        : Map<String, dynamic>.from(json);

    final ctxStr = raw['context']?.toString() ?? 'date';

    final ctx = InteractionContext.values.firstWhere(
      (e) => e.value == ctxStr,
      orElse: () => InteractionContext.date,
    );

    return Interaction(
      id: raw['id']?.toString() ?? '',
      fromUserId: raw['from_user_id']?.toString() ?? '',
      toUserId: raw['to_user_id']?.toString() ?? '',
      context: ctx,
      rating: (raw['rating'] ?? 3) as int,
      comment: raw['comment']?.toString(),
      confirmedAt: raw['confirmed_at'] != null
          ? DateTime.tryParse(raw['confirmed_at'].toString())
          : null,
    );
  }
}