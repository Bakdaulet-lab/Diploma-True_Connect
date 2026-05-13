import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

/// Background message handler — must be a top-level function.
@pragma('vm:entry-point')
Future<void> _firebaseMessagingBackgroundHandler(RemoteMessage message) async {
  await Firebase.initializeApp();
}

class PushService {
  PushService._();

  static final navigatorKey = GlobalKey<NavigatorState>();

  static String? _fcmToken;
  static String? get fcmToken => _fcmToken;

  static bool _initialized = false;

  static Future<void> init() async {
    if (_initialized) return;
    try {
      await Firebase.initializeApp();
      _initialized = true;
    } catch (_) {
      // Firebase not configured — skip push setup gracefully.
      return;
    }

    FirebaseMessaging.onBackgroundMessage(_firebaseMessagingBackgroundHandler);

    final messaging = FirebaseMessaging.instance;

    await messaging.requestPermission(
      alert: true,
      badge: true,
      sound: true,
    );

    _fcmToken = await messaging.getToken();

    // Refresh token events
    messaging.onTokenRefresh.listen((token) {
      _fcmToken = token;
    });

    // Foreground messages → in-app snackbar
    FirebaseMessaging.onMessage.listen(_handleForeground);

    // App opened from background notification tap
    FirebaseMessaging.onMessageOpenedApp.listen(_handleTap);

    // App launched from terminated state via notification tap
    final initial = await messaging.getInitialMessage();
    if (initial != null) _handleTap(initial);
  }

  static void _handleForeground(RemoteMessage message) {
    final context = navigatorKey.currentContext;
    if (context == null) return;

    final title = message.notification?.title ?? '';
    final body = message.notification?.body ?? '';
    final display = [title, body].where((s) => s.isNotEmpty).join(': ');

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(display),
        duration: const Duration(seconds: 4),
        behavior: SnackBarBehavior.floating,
        action: _buildAction(context, message),
      ),
    );
  }

  static SnackBarAction? _buildAction(
      BuildContext context, RemoteMessage message) {
    final route = _routeFor(message);
    if (route == null) return null;
    return SnackBarAction(
      label: 'Ашу',
      onPressed: () => context.push(route),
    );
  }

  static void _handleTap(RemoteMessage message) {
    final route = _routeFor(message);
    if (route == null) return;
    final context = navigatorKey.currentContext;
    if (context != null) {
      context.push(route);
    }
  }

  /// Maps notification data payload to an in-app route.
  static String? _routeFor(RemoteMessage message) {
    final data = message.data;
    final type = data['type'] as String?;
    final matchId = data['match_id'] as String?;
    final roomId = data['room_id'] as String?;

    switch (type) {
      case 'new_match':
        return '/matches';
      case 'new_message':
        if (matchId != null) return '/chat/$matchId';
        break;
      case 'mahram_message':
        if (roomId != null) return '/mahram-chat/$roomId';
        break;
      case 'niyyah_timer':
        return '/matches';
    }
    return null;
  }
}
