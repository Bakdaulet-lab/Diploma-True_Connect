import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../providers/feed_provider.dart';

class CreatePostScreen extends ConsumerStatefulWidget {
  const CreatePostScreen({super.key});

  @override
  ConsumerState<CreatePostScreen> createState() => _CreatePostScreenState();
}

class _CreatePostScreenState extends ConsumerState<CreatePostScreen> {
  final _contentCtrl = TextEditingController();
  bool _loading = false;
  String? _error;

  static const _maxChars = 2000;

  @override
  void dispose() {
    _contentCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final text = _contentCtrl.text.trim();
    if (text.isEmpty) {
      setState(() => _error = 'Мәтін жазыңыз');
      return;
    }
    setState(() {
      _loading = true;
      _error = null;
    });
    try {
      await ref.read(feedProvider.notifier).createPost(content: text);
      if (mounted) Navigator.of(context).pop();
    } catch (e) {
      setState(() => _error = e.toString());
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final remaining = _maxChars - _contentCtrl.text.length;
    final isOverLimit = remaining < 0;

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        title: Text(
          'Жаңа жазба',
          style: GoogleFonts.nunito(
            fontSize: 18,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
        ),
        actions: [
          Padding(
            padding:
                const EdgeInsets.only(right: AppSpacing.sm),
            child: TextButton(
              onPressed: _loading || isOverLimit ? null : _submit,
              child: _loading
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(
                          strokeWidth: 2, color: Colors.white),
                    )
                  : Text(
                      'Жариялау',
                      style: GoogleFonts.nunito(
                        fontSize: 14,
                        fontWeight: FontWeight.w700,
                        color: Colors.white,
                      ),
                    ),
            ),
          ),
        ],
      ),
      body: Padding(
        padding: const EdgeInsets.all(AppSpacing.md),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (_error != null)
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(AppSpacing.sm),
                margin: const EdgeInsets.only(bottom: AppSpacing.md),
                decoration: const BoxDecoration(
                  color: AppColors.accentLight,
                  borderRadius: AppRadius.input,
                ),
                child: Text(
                  _error!,
                  style: GoogleFonts.nunito(
                    fontSize: 13,
                    color: AppColors.accent,
                  ),
                ),
              ),
            Expanded(
              child: TextField(
                controller: _contentCtrl,
                maxLines: null,
                expands: true,
                maxLength: _maxChars,
                buildCounter: (_, {required currentLength, required isFocused, maxLength}) =>
                    null,
                textAlignVertical: TextAlignVertical.top,
                decoration: InputDecoration(
                  hintText:
                      'Ойыңызбен, пікіріңізбен немесе кеңесіңізбен бөлісіңіз...',
                  hintStyle: GoogleFonts.nunito(
                    fontSize: 15,
                    color: AppColors.textHint,
                    height: 1.6,
                  ),
                  border: InputBorder.none,
                  enabledBorder: InputBorder.none,
                  focusedBorder: InputBorder.none,
                  fillColor: Colors.transparent,
                ),
                style: GoogleFonts.nunito(
                  fontSize: 15,
                  color: AppColors.textPrimary,
                  height: 1.6,
                ),
                onChanged: (_) => setState(() {}),
              ),
            ),
            const SizedBox(height: AppSpacing.sm),
            Align(
              alignment: Alignment.centerRight,
              child: Text(
                '$remaining',
                style: GoogleFonts.nunito(
                  fontSize: 12,
                  color: isOverLimit ? AppColors.accent : AppColors.textHint,
                  fontWeight:
                      isOverLimit ? FontWeight.bold : FontWeight.normal,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
