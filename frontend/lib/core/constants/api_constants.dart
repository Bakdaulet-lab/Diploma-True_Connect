import 'dart:io';
import 'package:flutter/foundation.dart';

abstract final class ApiConstants {
  static String get baseUrl {
    // For development: use machine IP instead of localhost
    // to work around network isolation issues on Windows/web
    const ip = '172.22.192.1';
    return 'http://$ip:8080/v1';
  }

  static String get chatWs {
    const ip = '172.22.192.1';
    if (Platform.isAndroid) return 'ws://10.0.2.2:8080/v1/ws';
    return 'ws://$ip:8080/v1/ws';
  }

  static const timeout = Duration(seconds: 10);

  // Auth
  static const authRegister = '/auth/register';
  static const authLogin = '/auth/login';
  static const authRefresh = '/auth/refresh';
  static const authLogout = '/auth/logout';

  // Profile
  static const profile = '/users/me/profile';
  static const profilePhoto = '/users/me/photos';

  // Matching
  static const candidates = '/matching/candidates';
  static const swipe = '/matching/swipe';
  static const matches = '/matching/matches';

  // KYC
  static const kycSubmit = '/kyc/submit';

  // Settings
  static const settings = '/users/me/settings';

  // Imams
  static const imams = '/imams';

  // Mahram
  static const mahram = '/mahram';
  static const mahramRooms = '/mahram-rooms';

  // Whisper
  static const whisper = '/whisper';
}
