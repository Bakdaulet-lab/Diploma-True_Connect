import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/theme/app_colors.dart';
import '../../providers/notification_provider.dart';

class HomeShell extends ConsumerWidget {
  final StatefulNavigationShell navigationShell;

  const HomeShell({super.key, required this.navigationShell});

  void _onTabTap(int index) {
    // Tapping the already-active tab resets it to its branch root.
    navigationShell.goBranch(
      index,
      initialLocation: index == navigationShell.currentIndex,
    );
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final index = navigationShell.currentIndex;
    final unreadCount = ref.watch(unreadCountProvider).valueOrNull ?? 0;

    return Scaffold(
      body: navigationShell,
      bottomNavigationBar: Container(
        decoration: const BoxDecoration(
          color: AppColors.surface,
          border: Border(
            top: BorderSide(color: AppColors.divider, width: 0.5),
          ),
        ),
        child: NavigationBar(
          selectedIndex: index,
          onDestinationSelected: _onTabTap,
          backgroundColor: AppColors.surface,
          surfaceTintColor: Colors.transparent,
          indicatorColor: AppColors.primaryLight,
          elevation: 0,
          labelBehavior: NavigationDestinationLabelBehavior.alwaysShow,
          destinations: [
            _dest(
              icon: Icons.explore_outlined,
              activeIcon: Icons.explore,
              label: 'Табу',
            ),
            _dest(
              icon: Icons.article_outlined,
              activeIcon: Icons.article,
              label: 'Жаңалық',
            ),
            NavigationDestination(
              icon: Badge(
                isLabelVisible: unreadCount > 0,
                label: Text(
                  unreadCount > 99 ? '99+' : '$unreadCount',
                  style: const TextStyle(fontSize: 9),
                ),
                backgroundColor: AppColors.accent,
                textColor: Colors.white,
                child: const Icon(Icons.favorite_border),
              ),
              selectedIcon: Badge(
                isLabelVisible: unreadCount > 0,
                label: Text(
                  unreadCount > 99 ? '99+' : '$unreadCount',
                  style: const TextStyle(fontSize: 9),
                ),
                backgroundColor: AppColors.accent,
                textColor: Colors.white,
                child: const Icon(Icons.favorite, color: AppColors.primary),
              ),
              label: 'Сәйкестік',
            ),
            _dest(
              icon: Icons.person_outline,
              activeIcon: Icons.person,
              label: 'Профиль',
            ),
            _dest(
              icon: Icons.tune_outlined,
              activeIcon: Icons.tune,
              label: 'Баптаулар',
            ),
          ],
        ),
      ),
    );
  }

  NavigationDestination _dest({
    required IconData icon,
    required IconData activeIcon,
    required String label,
  }) {
    return NavigationDestination(
      icon: Icon(icon),
      selectedIcon: Icon(activeIcon, color: AppColors.primary),
      label: label,
    );
  }
}
