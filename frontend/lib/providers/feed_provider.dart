import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/models.dart';
import '../services/api_service.dart';

class FeedState {
  final List<Post> posts;
  final bool isLoading;
  final bool hasMore;
  final String? error;

  const FeedState({
    this.posts = const [],
    this.isLoading = false,
    this.hasMore = true,
    this.error,
  });

  FeedState copyWith({
    List<Post>? posts,
    bool? isLoading,
    bool? hasMore,
    String? error,
  }) =>
      FeedState(
        posts: posts ?? this.posts,
        isLoading: isLoading ?? this.isLoading,
        hasMore: hasMore ?? this.hasMore,
        error: error,
      );
}

class FeedNotifier extends StateNotifier<FeedState> {
  final ApiService _apiService;
  int _page = 1;

  FeedNotifier(this._apiService) : super(const FeedState());

  Future<void> loadPosts({bool refresh = false}) async {
    if (refresh) _page = 1;
    if (state.isLoading) return;

    state = state.copyWith(isLoading: true, error: null);
    try {
      final posts = await _apiService.getPosts(page: _page);
      state = state.copyWith(
        posts: refresh ? posts : [...state.posts, ...posts],
        isLoading: false,
        hasMore: posts.length >= 20,
      );
      _page++;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> createPost({
    required String content,
    String? mediaPath,
  }) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final post =
          await _apiService.createPost(content: content, mediaPath: mediaPath);
      state = state.copyWith(
        posts: [post, ...state.posts],
        isLoading: false,
      );
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> deletePost(String postId) async {
    try {
      await _apiService.deletePost(postId);
      state = state.copyWith(
        posts: state.posts.where((p) => p.id != postId).toList(),
      );
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }

  Future<void> toggleLike(String postId) async {
    final index = state.posts.indexWhere((p) => p.id == postId);
    if (index == -1) return;

    final post = state.posts[index];
    final updatedPost = post.copyWith(
      isLiked: !post.isLiked,
      likeCount: post.isLiked ? post.likeCount - 1 : post.likeCount + 1,
    );

    final updatedPosts = List<Post>.from(state.posts);
    updatedPosts[index] = updatedPost;
    state = state.copyWith(posts: updatedPosts);

    try {
      if (post.isLiked) {
        await _apiService.unlikePost(postId);
      } else {
        await _apiService.likePost(postId);
      }
    } catch (e) {
      // Revert on error
      final revertedPosts = List<Post>.from(state.posts);
      revertedPosts[index] = post;
      state = state.copyWith(posts: revertedPosts, error: e.toString());
    }
  }
}

final feedProvider = StateNotifierProvider<FeedNotifier, FeedState>((ref) {
  return FeedNotifier(ref.read(apiServiceProvider));
});

// Comments provider per post
final commentsProvider =
    FutureProvider.family<List<PostComment>, String>((ref, postId) async {
  return ref.read(apiServiceProvider).getComments(postId);
});
