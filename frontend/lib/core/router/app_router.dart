import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../providers/auth_provider.dart';
import '../services/push_service.dart';
import '../../screens/auth/login_screen.dart';
import '../../screens/auth/register_screen.dart';
import '../../screens/chat/call_screen.dart';
import '../../screens/chat/chat_screen.dart';
import '../../screens/chat/mahram_chat_screen.dart';
import '../../screens/discovery/discovery_screen.dart';
import '../../screens/feed/create_post_screen.dart';
import '../../screens/feed/feed_screen.dart';
import '../../screens/home/home_shell.dart';
import '../../screens/notifications/notifications_screen.dart';
import '../../services/websocket_service.dart';
import '../../screens/imams/imam_connect_screen.dart';
import '../../screens/kyc/kyc_screen.dart';
import '../../screens/matches/matches_screen.dart';
import '../../screens/niyyah/niyyah_selection_screen.dart';
import '../../screens/onboarding/onboarding_screen.dart';
import '../../screens/profile/profile_detail_screen.dart';
import '../../screens/profile/profile_screen.dart';
import '../../screens/matches/first_meeting_screen.dart';
import '../../screens/profile/edit_profile_screen.dart';
import '../../screens/profile/trust_breakdown_screen.dart';
import '../../screens/settings/settings_screen.dart';
import '../../screens/splash/splash_screen.dart';

/// Fires ChangeNotifier when auth state changes so GoRouter re-evaluates
/// redirects without recreating the GoRouter instance.
class _AuthRouterNotifier extends ChangeNotifier {
  _AuthRouterNotifier(Ref ref) {
    ref.listen<AsyncValue<dynamic>>(
      authStateProvider,
      (_, __) => notifyListeners(),
    );
  }
}

final _authRouterNotifierProvider = Provider<_AuthRouterNotifier>((ref) {
  return _AuthRouterNotifier(ref);
});

final routerProvider = Provider<GoRouter>((ref) {
  final notifier = ref.watch(_authRouterNotifierProvider);

  return GoRouter(
    navigatorKey: PushService.navigatorKey,
    initialLocation: '/splash',
    refreshListenable: notifier,
    redirect: (context, state) {
      final authState = ref.read(authStateProvider);
      final loc = state.matchedLocation;

      // Still restoring session — do not redirect yet.
      if (authState.isLoading) return null;

      final isAuthenticated = authState.valueOrNull != null;

      const publicRoutes = [
        '/splash',
        '/onboarding',
        '/auth/login',
        '/auth/register',
        '/niyyah',
      ];

      if (!isAuthenticated &&
          !publicRoutes.any((r) => loc.startsWith(r))) {
        return '/auth/login';
      }

      // Authenticated users should not stay on auth screens.
      if (isAuthenticated &&
          (loc == '/auth/login' || loc == '/auth/register')) {
        return '/home';
      }

      return null;
    },
    routes: [
      // ── Public ──────────────────────────────────────────────────────────
      GoRoute(
        path: '/splash',
        builder: (_, __) => const SplashScreen(),
      ),
      GoRoute(
        path: '/onboarding',
        builder: (_, __) => const OnboardingScreen(),
      ),
      GoRoute(
        path: '/niyyah',
        builder: (_, __) => const NiyyahSelectionScreen(),
      ),
      GoRoute(
        path: '/auth/login',
        builder: (_, __) => const LoginScreen(),
      ),
      GoRoute(
        path: '/auth/register',
        builder: (_, __) => const RegisterScreen(),
      ),

      // ── Shell (bottom nav) ───────────────────────────────────────────────
      ShellRoute(
        builder: (_, __, child) => HomeShell(child: child),
        routes: [
          GoRoute(
            path: '/home',
            builder: (_, __) => const DiscoveryScreen(),
          ),
          GoRoute(
            path: '/feed',
            builder: (_, __) => const FeedScreen(),
          ),
          GoRoute(
            path: '/matches',
            builder: (_, __) => const MatchesScreen(),
          ),
          GoRoute(
            path: '/profile',
            builder: (_, __) => const ProfileScreen(),
          ),
          GoRoute(
            path: '/settings',
            builder: (_, __) => const SettingsScreen(),
          ),
        ],
      ),

      // ── Feature screens ──────────────────────────────────────────────────
      GoRoute(
        path: '/feed/create',
        builder: (_, __) => const CreatePostScreen(),
      ),
      GoRoute(
        path: '/notifications',
        builder: (_, __) => const NotificationsScreen(),
      ),
      GoRoute(
        path: '/call/:matchId',
        builder: (_, state) {
          final extra = state.extra as Map<String, dynamic>? ?? {};
          return CallScreen(
            matchData: {'id': state.pathParameters['matchId'] ?? ''},
            wsService: extra['wsService'] as WebSocketService,
            isCaller: extra['isCaller'] as bool? ?? true,
            initialOffer: extra['initialOffer'] as Map<String, dynamic>?,
          );
        },
      ),
      GoRoute(
        path: '/chat/:matchId',
        builder: (_, state) {
          final matchId = state.pathParameters['matchId'] ?? '';
          if (matchId.isEmpty) {
            return const Scaffold(body: Center(child: Text('Match ID жоқ')));
          }
          return ChatScreen(
            matchId: matchId,
            otherUserId: state.uri.queryParameters['userId'] ?? '',
          );
        },
      ),
      GoRoute(
        path: '/mahram-chat/:roomId',
        builder: (_, state) => MahramChatScreen(
          roomId: state.pathParameters['roomId'] ?? '',
        ),
      ),

      // 1. Сначала точный маршрут
      GoRoute(
        path: '/profile/edit',
        builder: (_, __) => const EditProfileScreen(),
      ),
      GoRoute(
        path: '/profile/trust',
        builder: (_, __) => const TrustBreakdownScreen(),
      ),

      // 2. Затем маршрут для первой встречи (он не конфликтует)
      GoRoute(
        path: '/first-meeting/:matchId',
        builder: (_, state) {
          final matchId = state.pathParameters['matchId'] ?? '';
          if (matchId.isEmpty) {
            return const Scaffold(body: Center(child: Text('Match ID жоқ')));
          }
          return FirstMeetingScreen(
            matchId: matchId,
            otherUserId: state.uri.queryParameters['userId'] ?? '',
          );
        },
      ),

      // 3. Динамический маршрут В САМОМ КОНЦЕ
      GoRoute(
        path: '/profile/:userId',
        builder: (_, state) => ProfileDetailScreen(
          profile: {'id': state.pathParameters['userId'] ?? ''},
        ),
      ),
      GoRoute(
        path: '/kyc',
        builder: (_, __) => const KycScreen(),
      ),
      GoRoute(
        path: '/imams',
        builder: (_, state) => ImamConnectScreen(
          matchId: state.uri.queryParameters['matchId'],
        ),
      ),
    ],
    errorBuilder: (_, state) => Scaffold(
      body: Center(
        child: Text('Бет табылмады: ${state.error}'),
      ),
    ),
  );
});
