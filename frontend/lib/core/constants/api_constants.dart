import 'package:flutter/foundation.dart';
import 'platform_stub.dart' if (dart.library.io) 'platform_io.dart';

abstract final class ApiConstants {
  static String get _host {
    if (kIsWeb) return 'localhost';
    return isAndroid ? '10.0.2.2' : 'localhost';
  }

  static String get baseUrl {
    return 'http://$_host:8080/v1';
  }

  static String get chatWs {
    return 'ws://$_host:8080/v1/ws';
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

  // Social feed
  static const posts = '/posts';

  // Notifications
  static const notifications = '/notifications';

  // Chat history
  static String matchMessages(String matchId) => '/matches/$matchId/messages';

  // Push notifications
  static const fcmToken = '/users/me/fcm-token';
}
