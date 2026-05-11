import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
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
  bool _obscure = true;
  bool _loading = false;

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
          phone: _phoneCtr.text.trim(),
          password: _passCtr.text,
          name: _nameCtr.text.trim(),
        );

    if (!mounted) return;
    setState(() => _loading = false);

    final authState = ref.read(authStateProvider);
    authState.whenOrNull(
      error: (e, _) => ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text(e.toString()))),
      data: (user) {
        if (user != null) {
          // After registration → Niyyah selection (mandatory)
          context.go('/niyyah');
        }
      },
    );
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
                            decoration: const InputDecoration(
                              hintText: '+7 ___ ___ __ __',
                              prefixIcon: Icon(Icons.phone_outlined,
                                  color: AppColors.textHint),
                            ),
                            validator: (v) {
                              if (v == null || v.trim().isEmpty) {
                                return 'Телефон нөмірін енгізіңіз';
                              }
                              if (!v.trim().startsWith('+7') ||
                                  v.trim().length < 12) {
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

                          Text(
                            'Тіркелу арқылы сіз қызмет\nшарттарымен келісесіз',
                            textAlign: TextAlign.center,
                            style: GoogleFonts.nunito(
                              fontSize: 11,
                              color: AppColors.textHint,
                              height: 1.5,
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
