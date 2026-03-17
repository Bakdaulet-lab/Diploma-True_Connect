import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../providers/feed_provider.dart';
import '../../widgets/post_card.dart';
import '../../widgets/empty_state.dart';

class FeedScreen extends ConsumerStatefulWidget {
  const FeedScreen({super.key});

  @override
  ConsumerState<FeedScreen> createState() => _FeedScreenState();
}

class _FeedScreenState extends ConsumerState<FeedScreen> {
  final _scrollController = ScrollController();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(feedProvider.notifier).loadPosts(refresh: true);
    });
    _scrollController.addListener(_onScroll);
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  void _onScroll() {
    if (_scrollController.position.pixels >=
        _scrollController.position.maxScrollExtent - 200) {
      final state = ref.read(feedProvider);
      if (!state.isLoading && state.hasMore) {
        ref.read(feedProvider.notifier).loadPosts();
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(feedProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text(
          'Feed',
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () => context.push('/feed/create'),
        child: const Icon(Icons.add),
      ),
      body: _buildBody(state),
    );
  }

  Widget _buildBody(FeedState state) {
    if (state.isLoading && state.posts.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (state.posts.isEmpty) {
      return EmptyState(
        icon: Icons.dynamic_feed,
        title: 'No posts yet',
        subtitle: 'Be the first to share something!',
        actionLabel: 'Create Post',
        onAction: () => context.push('/feed/create'),
      );
    }

    return RefreshIndicator(
      onRefresh: () => ref.read(feedProvider.notifier).loadPosts(refresh: true),
      child: ListView.builder(
        controller: _scrollController,
        padding: const EdgeInsets.only(bottom: 80),
        itemCount: state.posts.length + (state.isLoading ? 1 : 0),
        itemBuilder: (context, index) {
          if (index == state.posts.length) {
            return const Center(
              child: Padding(
                padding: EdgeInsets.all(16),
                child: CircularProgressIndicator(),
              ),
            );
          }

          final post = state.posts[index];
          return PostCard(
            post: post,
            onLike: () =>
                ref.read(feedProvider.notifier).toggleLike(post.id),
            onDelete: () =>
                ref.read(feedProvider.notifier).deletePost(post.id),
          );
        },
      ),
    );
  }
}
