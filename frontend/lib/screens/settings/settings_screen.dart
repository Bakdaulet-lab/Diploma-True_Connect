import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/theme/app_colors.dart';
import '../../core/constants/app_strings.dart';
import '../../providers/settings_provider.dart';
import '../../providers/auth_provider.dart';

class SettingsScreen extends ConsumerStatefulWidget {
  const SettingsScreen({super.key});

  @override
  ConsumerState<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends ConsumerState<SettingsScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(settingsProvider.notifier).load();
    });
  }

  @override
  Widget build(BuildContext context) {
    final settingsState = ref.watch(settingsProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text(AppStrings.settingsTitle),
      ),
      body: settingsState.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error: $e')),
        data: (settings) => ListView(
          children: [
            const _SectionHeader(title: 'Discovery'),
            // Max Distance
            ListTile(
              leading: const Icon(Icons.social_distance),
              title: const Text(AppStrings.maxDistance),
              subtitle: Text('${settings.maxDistanceKm} km'),
              trailing: SizedBox(
                width: 180,
                child: Slider(
                  value: settings.maxDistanceKm.toDouble(),
                  min: 5,
                  max: 200,
                  divisions: 39,
                  label: '${settings.maxDistanceKm} km',
                  onChanged: (v) {
                    ref.read(settingsProvider.notifier).update({
                      'max_distance_km': v.round(),
                    });
                  },
                ),
              ),
            ),
            // Age Range
            ListTile(
              leading: const Icon(Icons.calendar_today),
              title: const Text(AppStrings.ageRange),
              subtitle: Text(
                  '${settings.ageRangeMin} - ${settings.ageRangeMax}'),
              trailing: SizedBox(
                width: 180,
                child: RangeSlider(
                  values: RangeValues(
                    settings.ageRangeMin.toDouble(),
                    settings.ageRangeMax.toDouble(),
                  ),
                  min: 18,
                  max: 80,
                  divisions: 62,
                  labels: RangeLabels(
                    settings.ageRangeMin.toString(),
                    settings.ageRangeMax.toString(),
                  ),
                  onChanged: (v) {
                    ref.read(settingsProvider.notifier).update({
                      'age_range_min': v.start.round(),
                      'age_range_max': v.end.round(),
                    });
                  },
                ),
              ),
            ),
            const Divider(),
            const _SectionHeader(title: 'Notifications'),
            SwitchListTile(
              secondary: const Icon(Icons.notifications_outlined),
              title: const Text(AppStrings.pushNotifications),
              value: settings.pushNotifications,
              onChanged: (v) {
                ref.read(settingsProvider.notifier).update({
                  'push_notifications': v,
                });
              },
            ),
            const Divider(),
            const _SectionHeader(title: 'Privacy'),
            SwitchListTile(
              secondary: const Icon(Icons.visibility_outlined),
              title: const Text(AppStrings.showOnlineStatus),
              value: settings.showOnlineStatus,
              onChanged: (v) {
                ref.read(settingsProvider.notifier).update({
                  'show_online_status': v,
                });
              },
            ),
            const Divider(),
            const _SectionHeader(title: 'Account'),
            ListTile(
              leading: const Icon(Icons.logout, color: AppColors.error),
              title: const Text(
                AppStrings.logOut,
                style: TextStyle(color: AppColors.error),
              ),
              onTap: () {
                showDialog(
                  context: context,
                  builder: (context) => AlertDialog(
                    title: const Text('Log Out'),
                    content: const Text(
                        'Are you sure you want to log out?'),
                    actions: [
                      TextButton(
                        onPressed: () => Navigator.pop(context),
                        child: const Text('Cancel'),
                      ),
                      ElevatedButton(
                        onPressed: () {
                          Navigator.pop(context);
                          ref.read(authProvider.notifier).logout();
                        },
                        style: ElevatedButton.styleFrom(
                          backgroundColor: AppColors.error,
                        ),
                        child: const Text('Log Out'),
                      ),
                    ],
                  ),
                );
              },
            ),
            ListTile(
              leading: const Icon(Icons.delete_forever,
                  color: AppColors.error),
              title: const Text(
                AppStrings.deleteAccount,
                style: TextStyle(color: AppColors.error),
              ),
              onTap: () {
                showDialog(
                  context: context,
                  builder: (context) => AlertDialog(
                    title: const Text('Delete Account'),
                    content: const Text(
                      'This action is irreversible. All your data will be permanently deleted.',
                    ),
                    actions: [
                      TextButton(
                        onPressed: () => Navigator.pop(context),
                        child: const Text('Cancel'),
                      ),
                      ElevatedButton(
                        onPressed: () {
                          Navigator.pop(context);
                          ref.read(authProvider.notifier).deleteAccount();
                        },
                        style: ElevatedButton.styleFrom(
                          backgroundColor: AppColors.error,
                        ),
                        child: const Text('Delete'),
                      ),
                    ],
                  ),
                );
              },
            ),
            const SizedBox(height: 40),
          ],
        ),
      ),
    );
  }
}

class _SectionHeader extends StatelessWidget {
  final String title;

  const _SectionHeader({required this.title});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(16, 16, 16, 8),
      child: Text(
        title.toUpperCase(),
        style: TextStyle(
          fontSize: 12,
          fontWeight: FontWeight.w600,
          color: Colors.grey[600],
          letterSpacing: 1.2,
        ),
      ),
    );
  }
}
