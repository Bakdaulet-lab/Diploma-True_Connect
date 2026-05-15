import '../core/constants/api_constants.dart';

class Post {
  final String id;
  final String authorId;
  final String content;
  final String? mediaUrl;
  final int likeCount;
  final int commentCount;
  final DateTime createdAt;
  final String authorName;
  final String? authorAvatarUrl;
  final bool isLiked;

  const Post({
    required this.id,
    required this.authorId,
    required this.content,
    this.mediaUrl,
    required this.likeCount,
    required this.commentCount,
    required this.createdAt,
    required this.authorName,
    this.authorAvatarUrl,
    this.isLiked = false,
  });

  factory Post.fromJson(Map<String, dynamic> json) {
    final data = json['data'] is Map
        ? Map<String, dynamic>.from(json['data'] as Map)
        : Map<String, dynamic>.from(json);

    return Post(
      id: _str(data, ['id']) ?? '',
      authorId: _str(data, ['author_id', 'authorId']) ?? '',
      content: _str(data, ['content']) ?? '',
      mediaUrl: ApiConstants.fixImageUrl(_str(data, ['media_url', 'mediaUrl'])),
      likeCount: _int(data, ['like_count', 'likeCount']) ?? 0,
      commentCount: _int(data, ['comment_count', 'commentCount']) ?? 0,
      createdAt: _parseDate(data['created_at'] ?? data['createdAt']),
      authorName: _str(data, ['author_name', 'authorName']) ?? 'Пайдаланушы',
      authorAvatarUrl: ApiConstants.fixImageUrl(
          _str(data, ['author_avatar', 'authorAvatar', 'author_avatar_url'])),
      isLiked:
          data['is_liked'] as bool? ?? data['isLiked'] as bool? ?? false,
    );
  }

  Post copyWith({
    int? likeCount,
    int? commentCount,
    bool? isLiked,
    String? content,
  }) =>
      Post(
        id: id,
        authorId: authorId,
        content: content ?? this.content,
        mediaUrl: mediaUrl,
        likeCount: likeCount ?? this.likeCount,
        commentCount: commentCount ?? this.commentCount,
        createdAt: createdAt,
        authorName: authorName,
        authorAvatarUrl: authorAvatarUrl,
        isLiked: isLiked ?? this.isLiked,
      );
}

class PostComment {
  final String id;
  final String postId;
  final String authorId;
  final String content;
  final DateTime createdAt;
  final String authorName;
  final String? authorAvatarUrl;

  const PostComment({
    required this.id,
    required this.postId,
    required this.authorId,
    required this.content,
    required this.createdAt,
    required this.authorName,
    this.authorAvatarUrl,
  });

  factory PostComment.fromJson(Map<String, dynamic> json) {
    final data = json['data'] is Map
        ? Map<String, dynamic>.from(json['data'] as Map)
        : Map<String, dynamic>.from(json);

    return PostComment(
      id: _str(data, ['id']) ?? '',
      postId: _str(data, ['post_id', 'postId']) ?? '',
      authorId: _str(data, ['author_id', 'authorId']) ?? '',
      content: _str(data, ['content']) ?? '',
      createdAt: _parseDate(data['created_at'] ?? data['createdAt']),
      authorName: _str(data, ['author_name', 'authorName']) ?? 'Пайдаланушы',
      authorAvatarUrl: ApiConstants.fixImageUrl(
          _str(data, ['author_avatar', 'authorAvatar'])),
    );
  }
}

String? _str(Map<String, dynamic> data, List<String> keys) {
  for (final key in keys) {
    final v = data[key];
    if (v is String && v.isNotEmpty) return v;
  }
  return null;
}

int? _int(Map<String, dynamic> data, List<String> keys) {
  for (final key in keys) {
    final v = data[key];
    if (v is int) return v;
    if (v is num) return v.toInt();
  }
  return null;
}

DateTime _parseDate(dynamic value) {
  if (value is String) {
    return DateTime.tryParse(value)?.toLocal() ?? DateTime.now();
  }
  return DateTime.now();
}
