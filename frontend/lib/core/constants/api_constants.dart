abstract final class ApiConstants {
  static const baseUrl = 'http://10.0.2.2:8080/v1'; // Android emulator → localhost
  static const timeout = Duration(seconds: 30);

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

  // Chat
  static const chatWs = 'ws://10.0.2.2:8080/v1/ws';

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
