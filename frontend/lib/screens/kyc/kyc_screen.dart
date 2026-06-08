import 'dart:async';
import 'dart:io';
import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:image_picker/image_picker.dart';
import '../../core/constants/api_constants.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../providers/auth_provider.dart';

enum _KycStatus { idle, uploading, polling, verified, rejected, banned }

class KycScreen extends ConsumerStatefulWidget {
  const KycScreen({super.key});

  @override
  ConsumerState<KycScreen> createState() => _KycScreenState();
}

class _KycScreenState extends ConsumerState<KycScreen> {
  XFile? _frontPhoto;
  XFile? _backPhoto;
  _KycStatus _status = _KycStatus.idle;
  String? _errorMessage;

  // Polls /kyc/status up to [maxAttempts] times with [interval] between each.
  // Resolves to true when the backend confirms photo_verified.
  Future<bool> _pollVerification(Dio dio,
      {int maxAttempts = 15, Duration interval = const Duration(seconds: 2)}) async {
    for (var i = 0; i < maxAttempts; i++) {
      await Future.delayed(interval);
      if (!mounted) return false;
      try {
        final resp = await dio.get('/kyc/status');
        final data = resp.data is Map ? (resp.data['data'] ?? resp.data) : null;
        if (data is Map) {
          final level = data['verification_level'] as String?;
          if (level == 'photo_verified') return true;
        }
      } catch (_) {
        // ignore transient network errors during polling
      }
    }
    return false;
  }

  Future<void> _pick(bool isFront) async {
    final picker = ImagePicker();
    final file = await picker.pickImage(source: ImageSource.gallery);
    if (file != null) {
      setState(() {
        if (isFront) {
          _frontPhoto = file;
        } else {
          _backPhoto = file;
        }
      });
    }
  }

  Future<void> _submit() async {
    if (_frontPhoto == null || _backPhoto == null) {
      setState(() => _errorMessage = 'Жеке куәліктің екі жағын таңдаңыз');
      return;
    }

    setState(() {
      _status = _KycStatus.uploading;
      _errorMessage = null;
    });

    final dio = ref.read(dioClientProvider).dio;

    try {
      final formData = FormData.fromMap({
        'document': await MultipartFile.fromFile(
          _frontPhoto!.path,
          filename: 'document.jpg',
        ),
      });

      await dio.post(
        ApiConstants.kycSubmit,
        data: formData,
        options: Options(contentType: 'multipart/form-data'),
      );
    } on DioException catch (e) {
      final data = e.response?.data;
      if (mounted) {
        setState(() {
          _status = _KycStatus.idle;
          _errorMessage = data is Map && data['error'] != null
              ? data['error']['message']
              : e.message ?? 'Ошибка загрузки';
        });
      }
      return;
    }

    // Document accepted (202). Now poll the status endpoint — the ML pipeline
    // runs synchronously in a goroutine and typically finishes within a few seconds.
    if (mounted) setState(() => _status = _KycStatus.polling);

    final verified = await _pollVerification(dio);
    if (!mounted) return;

    setState(() => _status = verified ? _KycStatus.verified : _KycStatus.rejected);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.primaryDark,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new, color: Colors.white, size: 18),
          onPressed: () => context.pop(),
        ),
        bottom: const PreferredSize(
          preferredSize: Size.fromHeight(1),
          child: Divider(height: 1, thickness: 1, color: AppColors.goldBorder),
        ),
        title: Text(
          'KYC Верификация',
          style: GoogleFonts.nunito(
            fontSize: 20,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
        ),
      ),
      body: switch (_status) {
        _KycStatus.polling => _buildPolling(),
        _KycStatus.verified => _buildVerified(),
        _KycStatus.rejected => _buildRejected(),
        _KycStatus.banned => _buildBanned(),
        _ => _buildForm(),
      },
    );
  }

  Widget _buildForm() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(AppSpacing.lg),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            padding: const EdgeInsets.all(AppSpacing.md),
            decoration: BoxDecoration(
              color: AppColors.primaryLight,
              borderRadius: AppRadius.card,
              border: Border.all(color: AppColors.primary.withValues(alpha: 0.3)),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Icon(Icons.shield_outlined, color: AppColors.primary, size: 24),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'Деректер қауіпсіздігі',
                        style: GoogleFonts.nunito(
                          fontSize: 14,
                          fontWeight: FontWeight.bold,
                          color: AppColors.primary,
                        ),
                      ),
                      const SizedBox(height: 4),
                      Text(
                        'Жеке куәлік деректері AES-256 шифрлаумен қорғалады '
                        'және тек верификация үшін пайдаланылады.',
                        style: GoogleFonts.nunito(
                            fontSize: 12,
                            color: AppColors.textSecondary,
                            height: 1.5),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),

          const SizedBox(height: AppSpacing.lg),

          Text(
            'Жеке куәлік суреттері',
            style: GoogleFonts.nunito(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          const SizedBox(height: AppSpacing.sm),
          Text(
            'Алдыңғы және артқы жағын таңдаңыз',
            style: GoogleFonts.nunito(fontSize: 13, color: AppColors.textSecondary),
          ),

          const SizedBox(height: AppSpacing.lg),

          _PhotoPicker(label: 'Алдыңғы жағы', file: _frontPhoto, onTap: () => _pick(true)),
          const SizedBox(height: AppSpacing.md),
          _PhotoPicker(label: 'Артқы жағы', file: _backPhoto, onTap: () => _pick(false)),

          if (_errorMessage != null) ...[
            const SizedBox(height: AppSpacing.md),
            Container(
              padding: const EdgeInsets.all(AppSpacing.sm),
              decoration: const BoxDecoration(
                color: AppColors.accentLight,
                borderRadius: AppRadius.chip,
              ),
              child: Text(
                _errorMessage!,
                style: GoogleFonts.nunito(fontSize: 13, color: AppColors.accent),
              ),
            ),
          ],

          const SizedBox(height: AppSpacing.xl),

          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed: _status == _KycStatus.uploading ? null : _submit,
              child: _status == _KycStatus.uploading
                  ? const SizedBox(
                      height: 20,
                      width: 20,
                      child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2),
                    )
                  : const Text('Жіберу'),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPolling() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const CircularProgressIndicator(color: AppColors.primary),
            const SizedBox(height: AppSpacing.lg),
            Text(
              'ML модель тексеруде...',
              style: GoogleFonts.nunito(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: AppSpacing.sm),
            Text(
              'Фотоңыздан жүз анықтауда. Бірнеше секунд күтіңіз.',
              textAlign: TextAlign.center,
              style: GoogleFonts.nunito(
                  fontSize: 14, color: AppColors.textSecondary, height: 1.6),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildVerified() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              padding: const EdgeInsets.all(AppSpacing.lg),
              decoration: BoxDecoration(
                color: AppColors.primary.withValues(alpha: 0.12),
                shape: BoxShape.circle,
              ),
              child: const Icon(Icons.verified_user_rounded,
                  size: 64, color: AppColors.primary),
            ),
            const SizedBox(height: AppSpacing.lg),
            Text(
              'Верификация сәтті өтті!',
              style: GoogleFonts.nunito(
                fontSize: 22,
                fontWeight: FontWeight.bold,
                color: AppColors.primary,
              ),
            ),
            const SizedBox(height: AppSpacing.sm),
            Text(
              'ML моделі жүзіңізді анықтап, верификацияны растады. '
              'Сіздің деңгейіңіз: photo_verified.',
              textAlign: TextAlign.center,
              style: GoogleFonts.nunito(
                  fontSize: 14, color: AppColors.textSecondary, height: 1.6),
            ),
            const SizedBox(height: AppSpacing.xl),
            ElevatedButton(
              onPressed: () => context.pop(),
              child: const Text('Жабу'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildRejected() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              padding: const EdgeInsets.all(AppSpacing.lg),
              decoration: BoxDecoration(
                color: Colors.orange.withValues(alpha: 0.12),
                shape: BoxShape.circle,
              ),
              child: const Icon(Icons.face_retouching_off_rounded,
                  size: 64, color: Colors.orange),
            ),
            const SizedBox(height: AppSpacing.lg),
            Text(
              'Жүз анықталмады',
              style: GoogleFonts.nunito(
                fontSize: 22,
                fontWeight: FontWeight.bold,
                color: Colors.orange.shade700,
              ),
            ),
            const SizedBox(height: AppSpacing.sm),
            Text(
              'ML модель фотодан жүзді анықтай алмады немесе сурет сапасы жеткіліксіз. '
              'Жарқырақ жерде, тура қарап, жаңа сурет жіберіп көріңіз.',
              textAlign: TextAlign.center,
              style: GoogleFonts.nunito(
                  fontSize: 14, color: AppColors.textSecondary, height: 1.6),
            ),
            const SizedBox(height: AppSpacing.xl),
            ElevatedButton(
              onPressed: () => setState(() {
                _status = _KycStatus.idle;
                _frontPhoto = null;
                _backPhoto = null;
                _errorMessage = null;
              }),
              child: const Text('Қайталап көру'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildBanned() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              padding: const EdgeInsets.all(AppSpacing.lg),
              decoration: BoxDecoration(
                color: AppColors.accent.withValues(alpha: 0.1),
                shape: BoxShape.circle,
              ),
              child: const Icon(Icons.block, size: 56, color: AppColors.accent),
            ),
            const SizedBox(height: AppSpacing.lg),
            Text(
              'Аккаунт бұғатталды',
              style: GoogleFonts.nunito(
                fontSize: 22,
                fontWeight: FontWeight.bold,
                color: AppColors.accent,
              ),
            ),
            const SizedBox(height: AppSpacing.sm),
            Text(
              'Бұл ЖСН бойынша аккаунт бұрын бұғатталған. '
              'TrueConnect саясаты бойынша жаңа аккаунт '
              'ашуға рұқсат берілмейді.',
              textAlign: TextAlign.center,
              style: GoogleFonts.nunito(
                  fontSize: 14, color: AppColors.textSecondary, height: 1.6),
            ),
            const SizedBox(height: AppSpacing.xl),
            OutlinedButton.icon(
              onPressed: () => context.pop(),
              style: OutlinedButton.styleFrom(
                foregroundColor: AppColors.accent,
                side: const BorderSide(color: AppColors.accent),
              ),
              icon: const Icon(Icons.arrow_back, size: 16),
              label: const Text('Артқа'),
            ),
          ],
        ),
      ),
    );
  }
}

// ─── Photo Picker ─────────────────────────────────────────────────────────────

class _PhotoPicker extends StatelessWidget {
  final String label;
  final XFile? file;
  final VoidCallback onTap;

  const _PhotoPicker({
    required this.label,
    required this.file,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        height: 140,
        decoration: BoxDecoration(
          color: file != null ? AppColors.primaryLight : AppColors.surfaceVariant,
          borderRadius: AppRadius.card,
          border: Border.all(
            color: file != null ? AppColors.primary : AppColors.divider,
            width: 1.5,
          ),
        ),
        child: file != null
            ? ClipRRect(
                borderRadius: AppRadius.card,
                child: Image.file(File(file!.path), fit: BoxFit.cover),
              )
            : Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.add_photo_alternate_outlined,
                      size: 36, color: AppColors.primary),
                  const SizedBox(height: 8),
                  Text(
                    label,
                    style: GoogleFonts.nunito(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: AppColors.primary,
                    ),
                  ),
                  Text(
                    'Суретті таңдаңыз',
                    style: GoogleFonts.nunito(fontSize: 12, color: AppColors.textHint),
                  ),
                ],
              ),
      ),
    );
  }
}
