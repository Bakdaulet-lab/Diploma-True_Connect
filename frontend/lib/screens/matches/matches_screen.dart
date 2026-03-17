import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:timeago/timeago.dart' as timeago;
import '../../core/theme/app_colors.dart';
import '../../providers/matching_provider.dart';
import '../../widgets/empty_state.dart';

class MatchesScreen extends ConsumerStatefulWidget {
  const MatchesScreen({super.key});

  @override
  ConsumerState<MatchesScreen> createState() => _MatchesScreenState();
}

class _MatchesScreenState extends ConsumerState<MatchesScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(matchingProvider.notifier).loadMatches(refresh: true);
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(matchingProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text(
          'Matches',
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
      ),
      body: _buildBody(state),
    );
  }

  Widget _buildBody(MatchingState state) {
    if (state.isLoading && state.matches.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (state.matches.isEmpty) {
      return const EmptyState(
        icon: Icons.favorite_border,
        title: 'No matches yet',
        subtitle: 'Start discovering people to find your match!',
      );
    }

    return RefreshIndicator(
      onRefresh: () =>
          ref.read(matchingProvider.notifier).loadMatches(refresh: true),
      child: ListView.builder(
        padding: const EdgeInsets.symmetric(vertical: 8),
        itemCount: state.matches.length,
        itemBuilder: (context, index) {
          final match = state.matches[index];
          final otherUser = match.otherUser;

          return ListTile(
            leading: CircleAvatar(
              radius: 28,
              backgroundColor: AppColors.primaryLight,
              backgroundImage: otherUser?.avatarUrl != null
                  ? NetworkImage(otherUser!.avatarUrl!)
                  : null,
              child: otherUser?.avatarUrl == null
                  ? Text(
                      (otherUser?.displayName ?? '?')[0].toUpperCase(),
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                      ),
                    )
                  : null,
            ),
            title: Text(
              otherUser?.displayName ?? 'Unknown',
              style: const TextStyle(fontWeight: FontWeight.w600),
            ),
            subtitle: Text(
              match.lastMessage?.content ?? 'Say hello! 👋',
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
              style: TextStyle(color: Colors.grey[600]),
            ),
            trailing: Text(
              match.matchedAt != null
                  ? timeago.format(match.matchedAt!)
                  : '',
              style: TextStyle(
                fontSize: 12,
                color: Colors.grey[500],
              ),
            ),
            onTap: () => context.push('/chat/${match.id}'),
          );
        },
      ),
    );
  }
}
