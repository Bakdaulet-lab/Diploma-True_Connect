import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/models.dart';
import '../services/api_service.dart';

final kycStatusProvider = FutureProvider<KycStatus>((ref) async {
  return ref.read(apiServiceProvider).getKycStatus();
});

final trustScoreProvider =
    FutureProvider.family<TrustScore, String>((ref, userId) async {
  return ref.read(apiServiceProvider).getReputation(userId);
});
