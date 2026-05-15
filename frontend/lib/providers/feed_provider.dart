import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../core/constants/api_constants.dart';
import '../core/network/dio_error_message.dart';
import '../models/post.dart';
import 'auth_provider.dart';

// ─── Feed (posts list) ────────────────────────────────────────────────────────

class FeedNotifier extends StateNotifier<AsyncValue<List<Post>>> {
  final Dio _dio;
  String? _cursor;
  bool _hasMore = true;
  bool _loadingMore = false;

  FeedNotifier(this._dio) : super(const AsyncValue.loading()) {
    load();
  }

  Future<void> load() async {
    state = const AsyncValue.loading();
    _cursor = null;
    _hasMore = true;
    try {
      final resp = await _dio.get(ApiConstants.posts);
      final list = _parsePosts(resp.data);
      _cursor = _extractCursor(resp.data);
      _hasMore = _cursor != null;
      state = AsyncValue.data(list);
    } on DioException catch (e, st) {
      state = AsyncValue.error(dioErrorMessage(e), st);
    } catch (e, st) {
      state = AsyncValue.error(e, st);
    }
  }

  Future<void> loadMore() async {
    if (!_hasMore || _cursor == null || _loadingMore) return;
    _loadingMore = true;
    final current = state.valueOrNull ?? [];
    try {
      final resp =
          await _dio.get(ApiConstants.posts, queryParameters: {'cursor': _cursor});
      final more = _parsePosts(resp.data);
      _cursor = _extractCursor(resp.data);
      _hasMore = _cursor != null && more.isNotEmpty;
      state = AsyncValue.data([...current, ...more]);
    } catch (_) {
      // Keep current state on load-more failure
    } finally {
      _loadingMore = false;
    }
  }

  Future<void> refresh() => load();

  void incrementCommentCount(String postId) {
    final posts = state.valueOrNull;
    if (posts == null) return;
    final idx = posts.indexWhere((p) => p.id == postId);
    if (idx == -1) return;
    final updated = List<Post>.from(posts);
    final p = updated[idx];
    updated[idx] = Post(
      id: p.id,
      authorId: p.authorId,
      content: p.content,
      mediaUrl: p.mediaUrl,
      likeCount: p.likeCount,
      commentCount: p.commentCount + 1,
      createdAt: p.createdAt,
      authorName: p.authorName,
      authorAvatarUrl: p.authorAvatarUrl,
      isLiked: p.isLiked,
    );
    state = AsyncValue.data(updated);
  }

  Future<void> likePost(String postId) async {
    final posts = state.valueOrNull;
    if (posts == null) return;
    final idx = posts.indexWhere((p) => p.id == postId);
    if (idx == -1) return;

    final post = posts[idx];
    // Optimistic update
    final updated = List<Post>.from(posts);
    updated[idx] = post.copyWith(
      isLiked: !post.isLiked,
      likeCount: post.isLiked ? post.likeCount - 1 : post.likeCount + 1,
    );
    state = AsyncValue.data(updated);

    try {
      if (!post.isLiked) {
        await _dio.post('${ApiConstants.posts}/$postId/like');
      } else {
        await _dio.delete('${ApiConstants.posts}/$postId/like');
      }
    } catch (_) {
      // Revert on failure
      final reverted = List<Post>.from(state.valueOrNull ?? []);
      final revertIdx = reverted.indexWhere((p) => p.id == postId);
      if (revertIdx != -1) reverted[revertIdx] = post;
      state = AsyncValue.data(reverted);
    }
  }

  Future<void> createPost({
    required String content,
    String? mediaUrl,
  }) async {
    try {
      await _dio.post(ApiConstants.posts, data: {
        'content': content,
        if (mediaUrl != null) 'media_url': mediaUrl,
      });
      // Refresh the full feed so the new post has author_name from the SQL JOIN.
      await load();
    } on DioException catch (e) {
      throw dioErrorMessage(e);
    }
  }
}

final feedProvider =
    StateNotifierProvider<FeedNotifier, AsyncValue<List<Post>>>((ref) {
  // Re-create (and reload) whenever the logged-in user changes.
  ref.watch(authStateProvider.select((s) => s.valueOrNull?.id));
  return FeedNotifier(ref.watch(dioClientProvider).dio);
});

// ─── Comments (per-post) ──────────────────────────────────────────────────────

final postCommentsProvider =
    FutureProvider.family<List<PostComment>, String>((ref, postId) async {
  final dio = ref.watch(dioClientProvider).dio;
  try {
    final resp = await dio.get('${ApiConstants.posts}/$postId/comments');
    final raw = resp.data is Map ? resp.data['data'] ?? resp.data : resp.data;
    if (raw is List) {
      return raw
          .whereType<Map>()
          .map((e) =>
              PostComment.fromJson(Map<String, dynamic>.from(e)))
          .toList();
    }
    return [];
  } on DioException catch (e) {
    throw dioErrorMessage(e);
  }
});

Future<void> addComment(
  WidgetRef ref,
  String postId,
  String content,
) async {
  final dio = ref.read(dioClientProvider).dio;
  await dio.post('${ApiConstants.posts}/$postId/comments',
      data: {'content': content});

  // Optimistically bump comment count on the feed card.
  ref.read(feedProvider.notifier).incrementCommentCount(postId);

  // Trigger a fresh fetch of the comment list.
  ref.invalidate(postCommentsProvider(postId));
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

List<Post> _parsePosts(dynamic payload) {
  final data = payload is Map ? payload['data'] ?? payload : payload;
  if (data is List) {
    return data
        .whereType<Map>()
        .map((e) => Post.fromJson(Map<String, dynamic>.from(e)))
        .toList();
  }
  if (data is Map) {
    final items = data['posts'] ?? data['items'];
    if (items is List) {
      return items
          .whereType<Map>()
          .map((e) => Post.fromJson(Map<String, dynamic>.from(e)))
          .toList();
    }
  }
  return [];
}

String? _extractCursor(dynamic payload) {
  if (payload is Map) {
    final meta = payload['meta'] ?? payload['cursor'];
    if (meta is Map) {
      final cursor = meta['next_cursor'] as String?;
      return (cursor != null && cursor.isNotEmpty) ? cursor : null;
    }
    if (meta is String && meta.isNotEmpty) return meta;
  }
  return null;
}
