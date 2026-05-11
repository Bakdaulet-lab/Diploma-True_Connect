import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../providers/auth_provider.dart';
import '../../widgets/halal_pattern_painter.dart';

class LoginScreen extends ConsumerStatefulWidget {
  const LoginScreen({super.key});

  @override
  ConsumerState<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends ConsumerState<LoginScreen> {
  final _form = GlobalKey<FormState>();
  final _phoneCtr = TextEditingController();
  final _passCtr = TextEditingController();
  bool _obscure = true;
  bool _loading = false;

  @override
  void dispose() {
    _phoneCtr.dispose();
    _passCtr.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    if (!_form.currentState!.validate()) return;
    setState(() => _loading = true);

    await ref.read(authStateProvider.notifier).login(
          phone: _phoneCtr.text.trim(),
          password: _passCtr.text,
        );

    if (!mounted) return;
    setState(() => _loading = false);

    final state = ref.read(authStateProvider);
    state.whenOrNull(
      error: (e, _) => _showError(e.toString()),
      data: (user) {
        if (user != null) context.go('/home');
      },
    );
  }

  void _showError(String msg) {
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text(msg)),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.primaryDark,
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
                crossAxisAlignment: CrossAxisAlignment.center,
                children: [
                  const SizedBox(height: AppSpacing.xxl),

                  const IslamicStarWidget(
                    size: 56,
                    color: AppColors.secondary,
                  ),
                  const SizedBox(height: AppSpacing.lg),

                  Text(
                    'TrueConnect',
                    style: GoogleFonts.nunito(
                      fontSize: 28,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                  Text(
                    'Mahabbat',
                    style: GoogleFonts.nunito(
                      fontSize: 12,
                      color: AppColors.secondary,
                      letterSpacing: 3,
                    ),
                  ),

                  const SizedBox(height: AppSpacing.xxl),

                  // Form card
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
                          Text(
                            'Кіру',
                            style: GoogleFonts.nunito(
                              fontSize: 22,
                              fontWeight: FontWeight.bold,
                              color: AppColors.textPrimary,
                            ),
                          ),
                          const SizedBox(height: AppSpacing.lg),

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
                              return null;
                            },
                          ),

                          const SizedBox(height: AppSpacing.md),

                          // Password
                          TextFormField(
                            controller: _passCtr,
                            obscureText: _obscure,
                            decoration: InputDecoration(
                              hintText: 'Құпия сөз',
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
                                : const Text('Кіру'),
                          ),

                          const SizedBox(height: AppSpacing.md),

                          Center(
                            child: TextButton(
                              onPressed: () => context.push('/auth/register'),
                              child: Text(
                                'Тіркелу',
                                style: GoogleFonts.nunito(
                                  fontSize: 14,
                                  fontWeight: FontWeight.w600,
                                  color: AppColors.primary,
                                ),
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
