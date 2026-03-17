import 'profile.dart';

class Match {
  final String id;
  final String userAId;
  final String userBId;
  final bool userALiked;
  final bool userBLiked;
  final DateTime? matchedAt;
  final DateTime createdAt;
  final Profile? otherUser;
  final Message? lastMessage;

  const Match({
    required this.id,
    required this.userAId,
    required this.userBId,
    this.userALiked = false,
    this.userBLiked = false,
    this.matchedAt,
    required this.createdAt,
    this.otherUser,
    this.lastMessage,
  });

  bool get isMatched => userALiked && userBLiked;

  factory Match.fromJson(Map<String, dynamic> json) => Match(
        id: json['id'] ?? '',
        userAId: json['user_a_id'] ?? '',
        userBId: json['user_b_id'] ?? '',
        userALiked: json['user_a_liked'] ?? false,
        userBLiked: json['user_b_liked'] ?? false,
        matchedAt: json['matched_at'] != null
            ? DateTime.parse(json['matched_at'])
            : null,
        createdAt: DateTime.parse(
            json['created_at'] ?? DateTime.now().toIso8601String()),
        otherUser: json['other_user'] != null
            ? Profile.fromJson(json['other_user'])
            : null,
        lastMessage: json['last_message'] != null
            ? Message.fromJson(json['last_message'])
            : null,
      );
}

class Message {
  final String id;
  final String matchId;
  final String senderId;
  final String content;
  final DateTime? readAt;
  final DateTime createdAt;

  const Message({
    required this.id,
    required this.matchId,
    required this.senderId,
    required this.content,
    this.readAt,
    required this.createdAt,
  });

  bool get isRead => readAt != null;

  factory Message.fromJson(Map<String, dynamic> json) => Message(
        id: json['id'] ?? '',
        matchId: json['match_id'] ?? '',
        senderId: json['sender_id'] ?? '',
        content: json['content'] ?? '',
        readAt: json['read_at'] != null
            ? DateTime.parse(json['read_at'])
            : null,
        createdAt: DateTime.parse(
            json['created_at'] ?? DateTime.now().toIso8601String()),
      );

  Map<String, dynamic> toJson() => {
        'match_id': matchId,
        'content': content,
      };
}
