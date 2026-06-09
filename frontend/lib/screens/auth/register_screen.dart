import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../core/constants/api_constants.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../core/utils/phone_formatter.dart';
import '../../providers/auth_provider.dart';
import '../../widgets/halal_pattern_painter.dart';

class RegisterScreen extends ConsumerStatefulWidget {
  const RegisterScreen({super.key});

  @override
  ConsumerState<RegisterScreen> createState() => _RegisterScreenState();
}

class _RegisterScreenState extends ConsumerState<RegisterScreen> {
  final _form = GlobalKey<FormState>();
  final _nameCtr = TextEditingController();
  final _phoneCtr = TextEditingController();
  final _passCtr = TextEditingController();
  String? _madhab;
  String? _gender;
  bool _obscure = true;
  bool _loading = false;

  static const _madhabOptions = [
    ('hanafi', 'Ханафи'),
    ('shafi', 'Шафии'),
    ('maliki', 'Маликий'),
    ('hanbali', 'Ханбали'),
  ];

  @override
  void dispose() {
    _nameCtr.dispose();
    _phoneCtr.dispose();
    _passCtr.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_form.currentState!.validate()) return;
    setState(() => _loading = true);

    await ref.read(authStateProvider.notifier).register(
          phone: _phoneCtr.text.replaceAll(' ', ''),
          password: _passCtr.text,
          name: _nameCtr.text.trim(),
        );

    if (!mounted) return;

    final authState = ref.read(authStateProvider);
    if (authState.hasError) {
      setState(() => _loading = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(authState.error.toString())),
      );
      return;
    }

    if (authState.valueOrNull != null && (_madhab != null || _gender != null)) {
      // Save gender + madhab to profile immediately after account creation.
      try {
        final dio = ref.read(dioClientProvider).dio;
        // Auto-set looking_for to the opposite gender so the discovery feed
        // immediately filters correctly without a separate settings step.
        final lookingFor = _gender == 'male'
            ? 'female'
            : _gender == 'female'
                ? 'male'
                : null;
        await dio.put(ApiConstants.profile, data: {
          'display_name': _nameCtr.text.trim(),
          if (_gender != null) 'gender': _gender,
          if (lookingFor != null) 'looking_for': lookingFor,
          if (_madhab != null) 'madhab': _madhab,
        });
      } catch (_) {
        // Non-blocking — user can update these later in profile edit.
      }
    }

    if (!mounted) return;
    setState(() => _loading = false);

    if (authState.valueOrNull != null) {
      context.go('/niyyah');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.primaryDark,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new,
              color: Colors.white, size: 18),
          onPressed: () => context.pop(),
        ),
      ),
      body: Stack(
        children: [
          const Positioned.fill(
            child: CustomPaint(
              painter: HalalPatternPainter(
                color: AppColors.primaryLight,
                opacity: 0.06,
              ),
            ),
          ),
          SafeArea(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(AppSpacing.xl),
              child: Column(
                children: [
                  Text(
                    'Тіркелу',
                    style: GoogleFonts.nunito(
                      fontSize: 26,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                  Text(
                    'TrueConnect-ке қош келдіңіз',
                    style: GoogleFonts.nunito(
                      fontSize: 14,
                      color: AppColors.secondary,
                    ),
                  ),

                  const SizedBox(height: AppSpacing.xl),

                  Container(
                    padding: const EdgeInsets.all(AppSpacing.lg),
                    decoration: const BoxDecoration(
                      color: AppColors.surface,
                      borderRadius: AppRadius.card,
                      boxShadow: AppShadows.elevated,
                    ),
                    child: Form(
                      key: _form,
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          // Name
                          TextFormField(
                            controller: _nameCtr,
                            textCapitalization: TextCapitalization.words,
                            decoration: const InputDecoration(
                              hintText: 'Аты-жөні',
                              prefixIcon: Icon(Icons.person_outline,
                                  color: AppColors.textHint),
                            ),
                            validator: (v) => (v == null || v.trim().isEmpty)
                                ? 'Аты-жөніңізді енгізіңіз'
                                : null,
                          ),
                          const SizedBox(height: AppSpacing.md),

                          // Phone
                          TextFormField(
                            controller: _phoneCtr,
                            keyboardType: TextInputType.phone,
                            inputFormatters: [KzPhoneFormatter()],
                            decoration: const InputDecoration(
                              hintText: '+7 ___ ___ __ __',
                              prefixIcon: Icon(Icons.phone_outlined,
                                  color: AppColors.textHint),
                            ),
                            validator: (v) {
                              final raw = (v ?? '').replaceAll(' ', '');
                              if (raw.isEmpty || raw == '+7') {
                                return 'Телефон нөмірін енгізіңіз';
                              }
                              if (!raw.startsWith('+7') || raw.length != 12) {
                                return 'Форматы: +7XXXXXXXXXX';
                              }
                              return null;
                            },
                          ),
                          const SizedBox(height: AppSpacing.md),

                          // Password
                          TextFormField(
                            controller: _passCtr,
                            obscureText: _obscure,
                            decoration: InputDecoration(
                              hintText: 'Құпия сөз (8+ символ)',
                              prefixIcon: const Icon(Icons.lock_outlined,
                                  color: AppColors.textHint),
                              suffixIcon: IconButton(
                                icon: Icon(
                                  _obscure
                                      ? Icons.visibility_outlined
                                      : Icons.visibility_off_outlined,
                                  color: AppColors.textHint,
                                ),
                                onPressed: () =>
                                    setState(() => _obscure = !_obscure),
                              ),
                            ),
                            validator: (v) {
                              if (v == null || v.isEmpty) {
                                return 'Құпия сөзді енгізіңіз';
                              }
                              if (v.length < 8) {
                                return 'Кемінде 8 символ';
                              }
                              return null;
                            },
                          ),
                          const SizedBox(height: AppSpacing.md),

                          // Gender (required for discovery to work correctly)
                          FormField<String>(
                            validator: (_) => null,
                            builder: (_) => Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Padding(
                                  padding: const EdgeInsets.only(
                                      left: 4, bottom: 8),
                                  child: Text(
                                    'Жынысыңыз',
                                    style: GoogleFonts.nunito(
                                      fontSize: 13,
                                      color: AppColors.textSecondary,
                                    ),
                                  ),
                                ),
                                Row(
                                  children: [
                                    _GenderChip(
                                      label: 'Ер',
                                      icon: Icons.male,
                                      selected: _gender == 'male',
                                      onTap: () =>
                                          setState(() => _gender = 'male'),
                                    ),
                                    const SizedBox(width: AppSpacing.sm),
                                    _GenderChip(
                                      label: 'Әйел',
                                      icon: Icons.female,
                                      selected: _gender == 'female',
                                      onTap: () =>
                                          setState(() => _gender = 'female'),
                                    ),
                                  ],
                                ),
                              ],
                            ),
                          ),
                          const SizedBox(height: AppSpacing.md),

                          // Madhab (optional)
                          DropdownButtonFormField<String>(
                            value: _madhab,
                            decoration: const InputDecoration(
                              hintText: 'Мазхаб (қосымша)',
                              prefixIcon: Icon(Icons.mosque_outlined,
                                  color: AppColors.textHint),
                            ),
                            dropdownColor: AppColors.surface,
                            items: _madhabOptions
                                .map((opt) => DropdownMenuItem(
                                      value: opt.$1,
                                      child: Text(
                                        opt.$2,
                                        style: GoogleFonts.nunito(
                                            color: AppColors.textPrimary),
                                      ),
                                    ))
                                .toList(),
                            onChanged: (v) => setState(() => _madhab = v),
                          ),

                          const SizedBox(height: AppSpacing.lg),

                          ElevatedButton(
                            onPressed: _loading ? null : _submit,
                            child: _loading
                                ? const SizedBox(
                                    height: 20,
                                    width: 20,
                                    child: CircularProgressIndicator(
                                      color: Colors.white,
                                      strokeWidth: 2,
                                    ),
                                  )
                                : const Text('Тіркелу'),
                          ),

                          const SizedBox(height: AppSpacing.sm),

                          Center(
                            child: Text(
                              'Тіркелу арқылы сіз қызмет\nшарттарымен келісесіз',
                              textAlign: TextAlign.center,
                              style: GoogleFonts.nunito(
                                fontSize: 11,
                                color: AppColors.textHint,
                                height: 1.5,
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _GenderChip extends StatelessWidget {
  final String label;
  final IconData icon;
  final bool selected;
  final VoidCallback onTap;

  const _GenderChip({
    required this.label,
    required this.icon,
    required this.selected,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return Expanded(
      child: GestureDetector(
        onTap: onTap,
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 180),
          padding: const EdgeInsets.symmetric(vertical: 12),
          decoration: BoxDecoration(
            color: selected ? AppColors.primary : AppColors.surfaceVariant,
            borderRadius: AppRadius.input,
            border: Border.all(
              color: selected ? AppColors.primary : AppColors.divider,
              width: selected ? 2 : 1,
            ),
          ),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(icon,
                  size: 18,
                  color: selected ? Colors.white : AppColors.textHint),
              const SizedBox(width: 6),
              Text(
                label,
                style: GoogleFonts.nunito(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: selected ? Colors.white : AppColors.textSecondary,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
