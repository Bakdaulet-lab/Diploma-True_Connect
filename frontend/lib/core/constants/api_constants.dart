import 'package:flutter/foundation.dart';
import 'platform_stub.dart' if (dart.library.io) 'platform_io.dart';

abstract final class ApiConstants {
  // Build-time overrides for production:
  //   flutter build ... --dart-define=API_HOST=api.trueconnect.kz \
  //                      --dart-define=API_SCHEME=https
  static const String _envHost = String.fromEnvironment('API_HOST');
  static const String _envScheme = String.fromEnvironment('API_SCHEME');

  static String get _host {
    if (_envHost.isNotEmpty) return _envHost;
    if (kIsWeb) return 'localhost';
    return isAndroid ? '10.0.2.2' : 'localhost';
  }

  static String get _scheme => _envScheme.isNotEmpty ? _envScheme : 'http';

  static String get _wsScheme => _scheme == 'https' ? 'wss' : 'ws';

  static String get baseUrl {
    return '$_scheme://$_host:8080/v1';
  }

  static String get chatWs {
    return '$_wsScheme://$_host:8080/v1/ws';
  }

  static const timeout = Duration(seconds: 10);

  // Auth
  static const authRegister = '/auth/register';
  static const authLogin = '/auth/login';
  static const authRefresh = '/auth/refresh';
  static const authLogout = '/auth/logout';

  // Profile
  static const profile = '/profiles/me';
  static const profiles = '/profiles';
  static const profilePhoto = '/profiles/me/photos';

  // Matching
  static const candidates = '/matching/candidates';
  static const matchLike = '/matching/like';
  static const matchPass = '/matching/pass';
  static const matches = '/matches';
  static const String interactions = '/interactions';

  // KYC
  static const kycSubmit = '/kyc/submit';

  // Settings
  static const settings = '/settings';

  // Imams
  static const imams = '/imams';

  // Mahram
  static const mahram = '/mahram';
  static const mahramRooms = '/mahram-rooms';

  // Whisper
  static const whisper = '/whisper';

  // Halal venues
  static const venues = '/venues';

  // Pending likes (users who liked you)
  static const pendingLikes = '/matching/likes';

  // Social feed
  static const posts = '/posts';

  // Notifications
  static const notifications = '/notifications';

  // Chat history
  static String matchMessages(String matchId) => '/matches/$matchId/messages';

  // Unmatch / Block
  static String unmatch(String matchId) => '/matches/$matchId/unmatch';
  static String blockUser(String userId) => '/users/$userId/block';

  // Push notifications
  static const fcmToken = '/users/me/fcm-token';

  // Fix image URLs: backend generates localhost:9000 URLs which are
  // unreachable from Android emulator (needs 10.0.2.2) or real devices.
  // Also handles bare object keys (e.g. "users/uuid/photos/uuid.jpg") returned
  // by some endpoints that don't pre-build a full URL on the backend.
  static String? fixImageUrl(String? url) {
    if (url == null || url.isEmpty) return null;
    if (!url.startsWith('http')) {
      final key = url.replaceAll(RegExp(r'^/+'), '');
      return '$_scheme://$_host:9000/trueconnect/$key';
    }
    return url
        .replaceFirst('localhost:9000', '$_host:9000')
        .replaceFirst('minio:9000', '$_host:9000');
  }
}
