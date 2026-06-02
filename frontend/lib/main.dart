import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';
import 'core/router/app_router.dart';
import 'core/services/push_service.dart';
import 'core/services/snack_bar_service.dart';
import 'core/theme/app_theme.dart';

void main() {
  WidgetsFlutterBinding.ensureInitialized();

  // Load fonts from bundled assets only — never fetch over the network at
  // runtime (that caused first-launch jank). Font files live in
  // assets/google_fonts/. If a file is missing, text falls back gracefully.
  GoogleFonts.config.allowRuntimeFetching = false;

  // Lock to portrait
  SystemChrome.setPreferredOrientations([
    DeviceOrientation.portraitUp,
    DeviceOrientation.portraitDown,
  ]);

  // Status bar: dark icons on light background
  SystemChrome.setSystemUIOverlayStyle(const SystemUiOverlayStyle(
    statusBarColor: Colors.transparent,
    statusBarIconBrightness: Brightness.dark,
  ));

  runApp(const ProviderScope(child: TrueConnectApp()));

  // Initialize push notifications AFTER the first frame. Doing it before
  // runApp() blocked startup on the OS permission dialog + FCM token network
  // call, making the app feel very slow to open. Fire-and-forget here.
  PushService.init();
}

class TrueConnectApp extends ConsumerWidget {
  const TrueConnectApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);
    return MaterialApp.router(
      title: 'TrueConnect',
      theme: AppTheme.light,
      routerConfig: router,
      scaffoldMessengerKey: scaffoldMessengerKey,
      debugShowCheckedModeBanner: false,
    );
  }
}
