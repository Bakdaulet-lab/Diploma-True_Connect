import 'profile.dart';

class Post {
  final String id;
  final String authorId;
  final String content;
  final String? mediaUrl;
  final int likeCount;
  final int commentCount;
  final bool isLiked;
  final Profile? author;
  final DateTime createdAt;
  final DateTime updatedAt;

  const Post({
    required this.id,
    required this.authorId,
    required this.content,
    this.mediaUrl,
    this.likeCount = 0,
    this.commentCount = 0,
    this.isLiked = false,
    this.author,
    required this.createdAt,
    required this.updatedAt,
  });

  factory Post.fromJson(Map<String, dynamic> json) => Post(
        id: json['id'] ?? '',
        authorId: json['author_id'] ?? '',
        content: json['content'] ?? '',
        mediaUrl: json['media_url'],
        likeCount: json['like_count'] ?? 0,
        commentCount: json['comment_count'] ?? 0,
        isLiked: json['is_liked'] ?? false,
        author:
            json['author'] != null ? Profile.fromJson(json['author']) : null,
        createdAt: DateTime.parse(
            json['created_at'] ?? DateTime.now().toIso8601String()),
        updatedAt: DateTime.parse(
            json['updated_at'] ?? DateTime.now().toIso8601String()),
      );

  Post copyWith({
    int? likeCount,
    int? commentCount,
    bool? isLiked,
  }) =>
      Post(
        id: id,
        authorId: authorId,
        content: content,
        mediaUrl: mediaUrl,
        likeCount: likeCount ?? this.likeCount,
        commentCount: commentCount ?? this.commentCount,
        isLiked: isLiked ?? this.isLiked,
        author: author,
        createdAt: createdAt,
        updatedAt: updatedAt,
      );
}

class PostComment {
  final String id;
  final String postId;
  final String authorId;
  final String content;
  final Profile? author;
  final DateTime createdAt;

  const PostComment({
    required this.id,
    required this.postId,
    required this.authorId,
    required this.content,
    this.author,
    required this.createdAt,
  });

  factory PostComment.fromJson(Map<String, dynamic> json) => PostComment(
        id: json['id'] ?? '',
        postId: json['post_id'] ?? '',
        authorId: json['author_id'] ?? '',
        content: json['content'] ?? '',
        author:
            json['author'] != null ? Profile.fromJson(json['author']) : null,
        createdAt: DateTime.parse(
            json['created_at'] ?? DateTime.now().toIso8601String()),
      );
}


