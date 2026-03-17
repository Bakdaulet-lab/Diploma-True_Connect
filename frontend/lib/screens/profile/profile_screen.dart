import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/theme/app_colors.dart';
import '../../providers/profile_provider.dart';
import '../../providers/auth_provider.dart';
import '../../widgets/photo_gallery.dart';

class ProfileScreen extends ConsumerStatefulWidget {
  const ProfileScreen({super.key});

  @override
  ConsumerState<ProfileScreen> createState() => _ProfileScreenState();
}

class _ProfileScreenState extends ConsumerState<ProfileScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(profileProvider.notifier).loadMyProfile();
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(profileProvider);
    final authState = ref.watch(authProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text(
          'Profile',
          style: TextStyle(fontWeight: FontWeight.bold),
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.settings_outlined),
            onPressed: () => context.push('/settings'),
          ),
        ],
      ),
      body: state.isLoading
          ? const Center(child: CircularProgressIndicator())
          : SingleChildScrollView(
              child: Column(
                children: [
                  const SizedBox(height: 20),
                  // Avatar
                  CircleAvatar(
                    radius: 60,
                    backgroundColor: AppColors.primaryLight,
                    backgroundImage: state.profile?.avatarUrl != null
                        ? NetworkImage(state.profile!.avatarUrl!)
                        : null,
                    child: state.profile?.avatarUrl == null
                        ? const Icon(Icons.person,
                            size: 60, color: Colors.white)
                        : null,
                  ),
                  const SizedBox(height: 16),
                  // Name
                  Text(
                    state.profile?.displayName ?? 'Set up your profile',
                    style: const TextStyle(
                      fontSize: 24,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  if (state.profile?.city.isNotEmpty == true)
                    Padding(
                      padding: const EdgeInsets.only(top: 4),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          const Icon(Icons.location_on,
                              size: 16, color: AppColors.textSecondaryLight),
                          const SizedBox(width: 4),
                          Text(
                            state.profile!.city,
                            style: const TextStyle(
                                color: AppColors.textSecondaryLight),
                          ),
                        ],
                      ),
                    ),
                  if (state.profile?.age != null)
                    Text(
                      '${state.profile!.age} years old',
                      style:
                          const TextStyle(color: AppColors.textSecondaryLight),
                    ),
                  const SizedBox(height: 8),
                  // Trust Score
                  if (authState.user != null)
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 16, vertical: 8),
                      decoration: BoxDecoration(
                        color: _trustColor(authState.user!.trustScore)
                            .withOpacity(0.1),
                        borderRadius: BorderRadius.circular(20),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.verified_user,
                            size: 18,
                            color: _trustColor(authState.user!.trustScore),
                          ),
                          const SizedBox(width: 6),
                          Text(
                            'Trust Score: ${authState.user!.trustScore}',
                            style: TextStyle(
                              fontWeight: FontWeight.w600,
                              color: _trustColor(authState.user!.trustScore),
                            ),
                          ),
                        ],
                      ),
                    ),
                  const SizedBox(height: 24),
                  // Edit Profile button
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 24),
                    child: OutlinedButton.icon(
                      onPressed: () => context.push('/profile/edit'),
                      icon: const Icon(Icons.edit),
                      label: const Text('Edit Profile'),
                    ),
                  ),
                  const SizedBox(height: 8),
                  Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 24),
                    child: OutlinedButton.icon(
                      onPressed: () => context.push('/kyc'),
                      icon: const Icon(Icons.badge_outlined),
                      label: const Text('Verify Identity'),
                    ),
                  ),
                  const SizedBox(height: 24),
                  // Bio
                  if (state.profile?.bio.isNotEmpty == true)
                    Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 24),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text(
                            'About',
                            style: TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          const SizedBox(height: 8),
                          Text(
                            state.profile!.bio,
                            style: const TextStyle(fontSize: 16, height: 1.5),
                          ),
                        ],
                      ),
                    ),
                  const SizedBox(height: 24),
                  // Photos
                  if (state.photos.isNotEmpty)
                    Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 24),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          const Text(
                            'Photos',
                            style: TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          const SizedBox(height: 12),
                          PhotoGallery(
                            photos: state.photos,
                            onDelete: (id) => ref
                                .read(profileProvider.notifier)
                                .deletePhoto(id),
                          ),
                        ],
                      ),
                    ),
                  const SizedBox(height: 40),
                ],
              ),
            ),
    );
  }

  Color _trustColor(int score) {
    if (score >= 80) return AppColors.trustHigh;
    if (score >= 50) return AppColors.trustMedium;
    return AppColors.trustLow;
  }
}
