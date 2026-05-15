import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/theme/app_colors.dart';
import '../../providers/notification_provider.dart';

class HomeShell extends ConsumerWidget {
  final Widget child;

  const HomeShell({super.key, required this.child});

  int _tabIndex(BuildContext context) {
    final loc = GoRouterState.of(context).matchedLocation;
    if (loc.startsWith('/feed')) return 1;
    if (loc.startsWith('/matches')) return 2;
    if (loc.startsWith('/profile')) return 3;
    if (loc.startsWith('/settings')) return 4;
    return 0;
  }

  void _onTabTap(BuildContext context, int index) {
    switch (index) {
      case 0:
        context.go('/home');
      case 1:
        context.go('/feed');
      case 2:
        context.go('/matches');
      case 3:
        context.go('/profile');
      case 4:
        context.go('/settings');
    }
  }

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final index = _tabIndex(context);
    final unreadCount = ref.watch(unreadCountProvider).valueOrNull ?? 0;

    return Scaffold(
      body: child,
      bottomNavigationBar: Container(
        decoration: const BoxDecoration(
          color: AppColors.surface,
          border: Border(
            top: BorderSide(color: AppColors.divider, width: 0.5),
          ),
        ),
        child: NavigationBar(
          selectedIndex: index,
          onDestinationSelected: (i) => _onTabTap(context, i),
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
