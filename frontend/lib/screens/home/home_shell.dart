import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../core/theme/app_colors.dart';

class HomeShell extends StatelessWidget {
  final Widget child;

  const HomeShell({super.key, required this.child});

  int _tabIndex(BuildContext context) {
    final loc = GoRouterState.of(context).matchedLocation;
    if (loc.startsWith('/matches')) return 1;
    if (loc.startsWith('/profile')) return 2;
    if (loc.startsWith('/settings')) return 3;
    return 0;
  }

  void _onTabTap(BuildContext context, int index) {
    switch (index) {
      case 0:
        context.go('/home');
      case 1:
        context.go('/matches');
      case 2:
        context.go('/profile');
      case 3:
        context.go('/settings');
    }
  }

  @override
  Widget build(BuildContext context) {
    final index = _tabIndex(context);
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
          items: const [
            BottomNavigationBarItem(
              icon: _NavIcon(icon: Icons.explore_outlined),
              activeIcon: _NavIcon(icon: Icons.explore, active: true),
              label: 'Табу',
            ),
            BottomNavigationBarItem(
              icon: _NavIcon(icon: Icons.favorite_border),
              activeIcon: _NavIcon(icon: Icons.favorite, active: true),
              label: 'Сәйкестік',
            ),
            BottomNavigationBarItem(
              icon: _NavIcon(icon: Icons.person_outline),
              activeIcon: _NavIcon(icon: Icons.person, active: true),
              label: 'Профиль',
            ),
            BottomNavigationBarItem(
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
