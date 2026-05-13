import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
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
          boxShadow: [
            BoxShadow(
              color: Color(0x12000000),
              blurRadius: 12,
              offset: Offset(0, -3),
            ),
          ],
        ),
        child: BottomNavigationBar(
          currentIndex: index,
          onTap: (i) => _onTabTap(context, i),
          elevation: 0,
          backgroundColor: AppColors.surface,
          selectedItemColor: AppColors.primary,
          unselectedItemColor: AppColors.textHint,
          selectedLabelStyle: GoogleFonts.nunito(
              fontSize: 11, fontWeight: FontWeight.w600),
          unselectedLabelStyle: GoogleFonts.nunito(fontSize: 11),
          type: BottomNavigationBarType.fixed,
          items: [
            const BottomNavigationBarItem(
              icon: _NavIcon(icon: Icons.explore_outlined),
              activeIcon: _NavIcon(icon: Icons.explore, active: true),
              label: 'Табу',
            ),
            const BottomNavigationBarItem(
              icon: _NavIcon(icon: Icons.article_outlined),
              activeIcon: _NavIcon(icon: Icons.article, active: true),
              label: 'Жаңалық',
            ),
            BottomNavigationBarItem(
              icon: _BadgedNavIcon(
                icon: Icons.favorite_border,
                count: unreadCount,
              ),
              activeIcon: _BadgedNavIcon(
                icon: Icons.favorite,
                count: unreadCount,
                active: true,
              ),
              label: 'Сәйкестік',
            ),
            const BottomNavigationBarItem(
              icon: _NavIcon(icon: Icons.person_outline),
              activeIcon: _NavIcon(icon: Icons.person, active: true),
              label: 'Профиль',
            ),
            const BottomNavigationBarItem(
              icon: _NavIcon(icon: Icons.tune_outlined),
              activeIcon: _NavIcon(icon: Icons.tune, active: true),
              label: 'Баптаулар',
            ),
          ],
        ),
      ),
    );
  }
}

class _NavIcon extends StatelessWidget {
  final IconData icon;
  final bool active;

  const _NavIcon({required this.icon, this.active = false});

  @override
  Widget build(BuildContext context) {
    return Icon(icon,
        size: 24,
        color: active ? AppColors.primary : AppColors.textHint);
  }
}

class _BadgedNavIcon extends StatelessWidget {
  final IconData icon;
  final int count;
  final bool active;

  const _BadgedNavIcon({
    required this.icon,
    required this.count,
    this.active = false,
  });

  @override
  Widget build(BuildContext context) {
    return Stack(
      clipBehavior: Clip.none,
      children: [
        Icon(icon,
            size: 24,
            color: active ? AppColors.primary : AppColors.textHint),
        if (count > 0)
          Positioned(
            top: -4,
            right: -6,
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
              decoration: const BoxDecoration(
                color: AppColors.accent,
                borderRadius: BorderRadius.all(Radius.circular(8)),
              ),
              child: Text(
                count > 99 ? '99+' : '$count',
                style: const TextStyle(
                  fontSize: 9,
                  color: Colors.white,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ),
      ],
    );
  }
}
