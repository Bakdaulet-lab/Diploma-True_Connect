import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';
import '../core/constants/api_constants.dart';
import '../core/theme/app_colors.dart';
import '../core/theme/app_theme.dart';
import '../providers/auth_provider.dart';
import '../widgets/halal_pattern_painter.dart';

/// Shows an anonymous whisper report bottom sheet.
/// Usage:
///   showWhisperModal(context, matchId: '...', reportedUserId: '...');
Future<void> showWhisperModal(
  BuildContext context, {
  required String matchId,
  required String reportedUserId,
}) {
  return showModalBottomSheet(
    context: context,
    isScrollControlled: true,
    backgroundColor: AppColors.surface,
    shape: const RoundedRectangleBorder(
      borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
    ),
    builder: (_) => _WhisperSheet(
      matchId: matchId,
      reportedUserId: reportedUserId,
    ),
  );
}

class _WhisperSheet extends ConsumerStatefulWidget {
  final String matchId;
  final String reportedUserId;

  const _WhisperSheet({
    required this.matchId,
    required this.reportedUserId,
  });

  @override
  ConsumerState<_WhisperSheet> createState() => _WhisperSheetState();
}

class _WhisperSheetState extends ConsumerState<_WhisperSheet> {
  final _feedbackCtr = TextEditingController();
  bool _loading = false;
  bool _sent = false;

  @override
  void dispose() {
    _feedbackCtr.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final text = _feedbackCtr.text.trim();
    if (text.isEmpty) return;

    setState(() => _loading = true);
    try {
      final dio = ref.read(dioClientProvider).dio;
      await dio.post(ApiConstants.whisper, data: {
        'reported_user_id': widget.reportedUserId,
        'match_id': widget.matchId,
        'feedback': text,
      });
      if (mounted) setState(() => _sent = true);
    } catch (_) {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.fromLTRB(
        AppSpacing.lg,
        AppSpacing.lg,
        AppSpacing.lg,
        AppSpacing.xl + MediaQuery.of(context).viewInsets.bottom,
      ),
      child: _sent ? _buildConfirmation() : _buildForm(),
    );
  }

  Widget _buildForm() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Header
        Row(
          children: [
            const Icon(Icons.report_gmailerrorred_outlined,
                color: AppColors.accent, size: 22),
            const SizedBox(width: 8),
            Text(
              'Анонимді пікір',
              style: GoogleFonts.nunito(
                fontSize: 20,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
          ],
        ),
        const SizedBox(height: 4),
        Text(
          'Пікіріңіз анонимді жіберіледі. Жеке деректеріңіз ашылмайды.',
          style: GoogleFonts.nunito(
              fontSize: 13, color: AppColors.textSecondary, height: 1.4),
        ),

        const SizedBox(height: AppSpacing.lg),

        // Feedback text field
        TextField(
          controller: _feedbackCtr,
          maxLines: 4,
          maxLength: 500,
          decoration: InputDecoration(
            hintText: 'Сіз не байқадыңыз? Нақты және дұрыс жазыңыз...',
            hintStyle:
                GoogleFonts.nunito(fontSize: 13, color: AppColors.textHint),
            filled: true,
            fillColor: AppColors.surfaceVariant,
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(12),
              borderSide: BorderSide.none,
            ),
            counterStyle:
                GoogleFonts.nunito(fontSize: 11, color: AppColors.textHint),
          ),
          style:
              GoogleFonts.nunito(fontSize: 14, color: AppColors.textPrimary),
          textCapitalization: TextCapitalization.sentences,
        ),

        const SizedBox(height: AppSpacing.sm),

        // Warning note
        Container(
          padding: const EdgeInsets.all(AppSpacing.sm),
          decoration: BoxDecoration(
            color: AppColors.accentLight,
            borderRadius: AppRadius.chip,
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Icon(Icons.info_outline,
                  size: 14, color: AppColors.accent),
              const SizedBox(width: 6),
              Expanded(
                child: Text(
                  '3 немесе одан көп шынайы шағым жасалса, '
                  'аккаунт администраторға жіберіледі.',
                  style: GoogleFonts.nunito(
                      fontSize: 11, color: AppColors.accent, height: 1.4),
                ),
              ),
            ],
          ),
        ),

        const SizedBox(height: AppSpacing.lg),

        Row(
          children: [
            Expanded(
              child: OutlinedButton(
                onPressed: () => Navigator.pop(context),
                style: OutlinedButton.styleFrom(
                  foregroundColor: AppColors.textSecondary,
                  side: const BorderSide(color: AppColors.divider),
                ),
                child: const Text('Бас тарту'),
              ),
            ),
            const SizedBox(width: AppSpacing.md),
            Expanded(
              flex: 2,
              child: ElevatedButton(
                onPressed: _loading ? null : _submit,
                style: ElevatedButton.styleFrom(
                  backgroundColor: AppColors.accent,
                ),
                child: _loading
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
      ],
    );
  }

  Widget _buildConfirmation() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        const SizedBox(height: AppSpacing.lg),
        const IslamicStarWidget(size: 48, color: AppColors.primary),
        const SizedBox(height: AppSpacing.md),
        Text(
          'Пікір жіберілді',
          style: GoogleFonts.nunito(
            fontSize: 20,
            fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
          ),
        ),
        const SizedBox(height: 8),
        Text(
          'Пікіріңіз анонимді жіберілді.\nРахмет — TrueConnect қауымдастығын қауіпсіз ұстауға көмектестіңіз.',
          textAlign: TextAlign.center,
          style: GoogleFonts.nunito(
              fontSize: 13, color: AppColors.textSecondary, height: 1.5),
        ),
        const SizedBox(height: AppSpacing.xl),
        SizedBox(
          width: double.infinity,
          child: ElevatedButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Жабу'),
          ),
        ),
        const SizedBox(height: AppSpacing.md),
      ],
    );
  }
}
