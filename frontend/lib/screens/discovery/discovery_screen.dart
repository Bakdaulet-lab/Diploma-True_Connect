import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/theme/app_colors.dart';
import '../../providers/matching_provider.dart';
import '../../widgets/profile_card.dart';
import '../../widgets/empty_state.dart';

class DiscoveryScreen extends ConsumerStatefulWidget {
  const DiscoveryScreen({super.key});

  @override
  ConsumerState<DiscoveryScreen> createState() => _DiscoveryScreenState();
}

class _DiscoveryScreenState extends ConsumerState<DiscoveryScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(matchingProvider.notifier).loadCandidates(refresh: true);
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(matchingProvider);

    // Show match dialog
    ref.listen<MatchingState>(matchingProvider, (prev, next) {
      if (next.newMatch != null) {
        _showMatchDialog(context);
        ref.read(matchingProvider.notifier).clearNewMatch();
      }
    });

    return Scaffold(
      appBar: AppBar(
        title: const Text(
          'Discover',
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.tune),
            onPressed: () => Navigator.pushNamed(context, '/settings'),
          ),
        ],
      ),
      body: _buildBody(state),
    );
  }

  Widget _buildBody(MatchingState state) {
    if (state.isLoading && state.candidates.isEmpty) {
      return const Center(child: CircularProgressIndicator());
    }

    if (state.candidates.isEmpty) {
      return EmptyState(
        icon: Icons.explore_off,
        title: 'No more candidates',
        subtitle: 'Try adjusting your distance and age filters',
        actionLabel: 'Refresh',
        onAction: () =>
            ref.read(matchingProvider.notifier).loadCandidates(refresh: true),
      );
    }

    final candidate = state.candidates.first;

    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        children: [
          Expanded(
            child: ProfileCard(profile: candidate),
          ),
          const SizedBox(height: 20),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceEvenly,
            children: [
              // Pass button
              _ActionButton(
                icon: Icons.close,
                color: AppColors.error,
                size: 64,
                onTap: () => ref
                    .read(matchingProvider.notifier)
                    .passUser(candidate.userId),
              ),
              // Like button
              _ActionButton(
                icon: Icons.favorite,
                color: AppColors.secondary,
                size: 72,
                onTap: () => ref
                    .read(matchingProvider.notifier)
                    .likeUser(candidate.userId),
              ),
            ],
          ),
          const SizedBox(height: 20),
        ],
      ),
    );
  }

  void _showMatchDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const Icon(
              Icons.favorite,
              size: 64,
              color: AppColors.secondary,
            ),
            const SizedBox(height: 16),
            const Text(
              "It's a Match!",
              style: TextStyle(
                fontSize: 24,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'You can now message each other',
              style: TextStyle(color: Colors.grey[600]),
            ),
            const SizedBox(height: 24),
            ElevatedButton(
              onPressed: () => Navigator.pop(context),
              child: const Text('Keep Swiping'),
            ),
            TextButton(
              onPressed: () {
                Navigator.pop(context);
                // Navigate to matches
              },
              child: const Text('Send a Message'),
            ),
          ],
        ),
      ),
    );
  }
}

class _ActionButton extends StatelessWidget {
  final IconData icon;
  final Color color;
  final double size;
  final VoidCallback onTap;

  const _ActionButton({
    required this.icon,
    required this.color,
    required this.size,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: size,
        height: size,
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          color: Colors.white,
          boxShadow: [
            BoxShadow(
              color: color.withOpacity(0.3),
              blurRadius: 12,
              offset: const Offset(0, 4),
            ),
          ],
          border: Border.all(color: color.withOpacity(0.3), width: 2),
        ),
        child: Icon(icon, color: color, size: size * 0.45),
      ),
    );
  }
}
