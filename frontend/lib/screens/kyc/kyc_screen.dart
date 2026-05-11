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
import '../../widgets/halal_pattern_painter.dart';

enum _KycStatus { idle, uploading, pending, verified }

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
      setState(() =>
          _errorMessage = 'Жеке куәліктің екі жағын таңдаңыз');
      return;
    }

    setState(() {
      _status = _KycStatus.uploading;
      _errorMessage = null;
    });

    try {
      final dio = ref.read(dioClientProvider).dio;
      final formData = FormData.fromMap({
        'front': await MultipartFile.fromFile(_frontPhoto!.path,
            filename: 'front.jpg'),
        'back': await MultipartFile.fromFile(_backPhoto!.path,
            filename: 'back.jpg'),
      });
      await dio.post(ApiConstants.kycSubmit, data: formData);
      if (mounted) setState(() => _status = _KycStatus.pending);
    } catch (e) {
      if (mounted) {
        setState(() {
          _status = _KycStatus.idle;
          _errorMessage = 'Жүктеу қатесі: ${e.toString()}';
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.primaryDark,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new,
              color: Colors.white, size: 18),
          onPressed: () => context.pop(),
        ),
        bottom: const PreferredSize(
          preferredSize: Size.fromHeight(1),
          child:
              Divider(height: 1, thickness: 1, color: AppColors.goldBorder),
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
      body: _status == _KycStatus.pending || _status == _KycStatus.verified
          ? _buildSuccess()
          : _buildForm(),
    );
  }

  Widget _buildForm() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(AppSpacing.lg),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Info card
          Container(
            padding: const EdgeInsets.all(AppSpacing.md),
            decoration: BoxDecoration(
              color: AppColors.primaryLight,
              borderRadius: AppRadius.card,
              border:
                  Border.all(color: AppColors.primary.withValues(alpha: 0.3)),
            ),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Icon(Icons.shield_outlined,
                    color: AppColors.primary, size: 24),
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
            style: GoogleFonts.nunito(
                fontSize: 13, color: AppColors.textSecondary),
          ),

          const SizedBox(height: AppSpacing.lg),

          // Front photo
          _PhotoPicker(
            label: 'Алдыңғы жағы',
            file: _frontPhoto,
            onTap: () => _pick(true),
          ),

          const SizedBox(height: AppSpacing.md),

          // Back photo
          _PhotoPicker(
            label: 'Артқы жағы',
            file: _backPhoto,
            onTap: () => _pick(false),
          ),

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
                style: GoogleFonts.nunito(
                    fontSize: 13, color: AppColors.accent),
              ),
            ),
          ],

          const SizedBox(height: AppSpacing.xl),

          SizedBox(
            width: double.infinity,
            child: ElevatedButton(
              onPressed:
                  _status == _KycStatus.uploading ? null : _submit,
              child: _status == _KycStatus.uploading
                  ? const SizedBox(
                      height: 20,
                      width: 20,
                      child: CircularProgressIndicator(
                          color: Colors.white, strokeWidth: 2),
                    )
                  : const Text('Жіберу'),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSuccess() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const IslamicStarWidget(size: 64, color: AppColors.secondary),
            const SizedBox(height: AppSpacing.lg),
            Text(
              'Тексерілуде',
              style: GoogleFonts.nunito(
                fontSize: 24,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            const SizedBox(height: AppSpacing.sm),
            Text(
              'Құжаттарыңыз жіберілді. Верификация нәтижесі '
              '1-2 жұмыс күні ішінде хабарланады.',
              textAlign: TextAlign.center,
              style: GoogleFonts.nunito(
                  fontSize: 14,
                  color: AppColors.textSecondary,
                  height: 1.6),
            ),
            const SizedBox(height: AppSpacing.xl),
            ElevatedButton(
              onPressed: () => context.pop(),
              child: const Text('Артқа'),
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
          color: file != null
              ? AppColors.primaryLight
              : AppColors.surfaceVariant,
          borderRadius: AppRadius.card,
          border: Border.all(
            color: file != null
                ? AppColors.primary
                : AppColors.divider,
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
                    style: GoogleFonts.nunito(
                        fontSize: 12, color: AppColors.textHint),
                  ),
                ],
              ),
      ),
    );
  }
}
