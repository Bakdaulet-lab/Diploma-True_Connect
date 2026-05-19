import 'package:flutter/foundation.dart';

/// Bumped whenever the network layer detects an unrecoverable auth failure
/// (the refresh token was rejected — expired, rotated, or reused). The auth
/// layer listens to this and forces a logged-out state so the router can
/// redirect to /auth/login. Kept as a top-level signal (mirroring
/// scaffoldMessengerKey) so the Dio layer stays decoupled from Riverpod.
final sessionExpiredNotifier = ValueNotifier<int>(0);

void notifySessionExpired() {
  sessionExpiredNotifier.value++;
}
