import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:dio/dio.dart';
import '../../core/constants/api_constants.dart';
import '../../core/network/dio_error_message.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../models/profile.dart';
import '../../providers/auth_provider.dart';
import '../../widgets/halal_pattern_painter.dart';
import '../../widgets/niyyah_badge.dart';

// ─── Provider ─────────────────────────────────────────────────────────────────

final _editProfileProvider = FutureProvider<Profile>((ref) async {
  final dio = ref.watch(dioClientProvider).dio;
  try {
    final resp = await dio.get(ApiConstants.profile);
    return Profile.fromJson(resp.data as Map<String, dynamic>);
  } on DioException catch (e) {
    throw dioErrorMessage(e);
  }
});

// ─── Screen ───────────────────────────────────────────────────────────────────

class EditProfileScreen extends ConsumerWidget {
  const EditProfileScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncProfile = ref.watch(_editProfileProvider);

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.primaryDark,
        centerTitle: true,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios_new,
              color: Colors.white, size: 18),
          onPressed: () => context.pop(),
        ),
        bottom: const PreferredSize(
          preferredSize: Size.fromHeight(1),
          child: Divider(height: 1, thickness: 1, color: AppColors.goldBorder),
        ),
        title: Text(
          'Профильді өңдеу',
          style: GoogleFonts.nunito(
            fontSize: 20,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
        ),
      ),
      body: asyncProfile.when(
        loading: () => const Center(
            child: CircularProgressIndicator(color: AppColors.primary)),
        error: (e, _) => _buildError(context, ref, e.toString()),
        data: (profile) => _EditForm(profile: profile),
      ),
    );
  }

  Widget _buildError(BuildContext context, WidgetRef ref, String msg) {
    final isMissing =
        msg.contains('NOT_FOUND') || msg.contains('resource not found');
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.lg),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              isMissing ? Icons.person_add_alt_1 : Icons.error_outline,
              size: 48,
              color: AppColors.textHint,
            ),
            const SizedBox(height: 16),
            Text(
              isMissing ? 'Профиль табылмады' : msg,
              style: GoogleFonts.nunito(color: AppColors.textSecondary),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: () => ref.refresh(_editProfileProvider.future),
              child: const Text('Қайтадан'),
            ),
          ],
        ),
      ),
    );
  }
}

// ─── Form ─────────────────────────────────────────────────────────────────────

class _EditForm extends ConsumerStatefulWidget {
  final Profile profile;
  const _EditForm({required this.profile});

  @override
  ConsumerState<_EditForm> createState() => _EditFormState();
}

class _EditFormState extends ConsumerState<_EditForm> {
  final _formKey = GlobalKey<FormState>();
  late final TextEditingController _nameCtr;
  late final TextEditingController _cityCtr;
  late final TextEditingController _bioCtr;
  late final TextEditingController _ageCtr;

  late String? _niyyah;
  late String? _madhab;
  late Set<String> _languages;
  late bool _noPhotoMode;
  bool _saving = false;
  String? _error;

  static const _niyyahOptions = [
    ('nikah_year', '🌙', 'Никях'),
    ('serious_marriage', '💍', 'Серьёзно'),
    ('friendship', '🤝', 'Танысу'),
  ];

  static const _madhabOptions = [
    ('hanafi', 'Ханафи'),
    ('shafi', 'Шафии'),
    ('maliki', 'Маликий'),
    ('hanbali', 'Ханбали'),
  ];

  static const _languageOptions = [
    ('kaz', 'Қазақша'),
    ('rus', 'Русский'),
    ('ara', 'عربي'),
  ];

  @override
  void initState() {
    super.initState();
    final p = widget.profile;
    _nameCtr = TextEditingController(text: p.displayName);
    _cityCtr = TextEditingController(text: p.city ?? '');
    _bioCtr = TextEditingController(text: p.bio ?? '');
    _ageCtr =
        TextEditingController(text: p.age != null ? p.age.toString() : '');
    _niyyah = p.niyyah;
    _madhab = p.madhab;
    _languages = Set.from(p.languages);
    _noPhotoMode = p.noPhotoMode;
  }

  @override
  void dispose() {
    _nameCtr.dispose();
    _cityCtr.dispose();
    _bioCtr.dispose();
    _ageCtr.dispose();
    super.dispose();
  }

  Future<void> _save() async {
    if (!(_formKey.currentState?.validate() ?? false)) return;
    setState(() {
      _saving = true;
      _error = null;
    });

    try {
      final dio = ref.read(dioClientProvider).dio;
      final ageVal = int.tryParse(_ageCtr.text.trim());
      await dio.put(ApiConstants.profile, data: {
        'display_name': _nameCtr.text.trim(),
        if (_cityCtr.text.trim().isNotEmpty) 'city': _cityCtr.text.trim(),
        if (_bioCtr.text.trim().isNotEmpty) 'bio': _bioCtr.text.trim(),
        if (ageVal != null) 'age': ageVal,
        if (_niyyah != null) 'niyyah': _niyyah,
        if (_madhab != null) 'madhab': _madhab,
        'languages': _languages.toList(),
        'no_photo_mode': _noPhotoMode,
      });

      if (!mounted) return;
      ref.invalidate(_editProfileProvider);
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Профиль сақталды'),
          backgroundColor: AppColors.primary,
        ),
      );
      context.pop();
    } on DioException catch (e) {
      if (!mounted) return;
      setState(() {
        _saving = false;
        _error = dioErrorMessage(e);
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _saving = false;
        _error = e.toString();
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Stack(
      children: [
        const Positioned.fill(
          child: CustomPaint(
            painter: HalalPatternPainter(
              color: AppColors.primaryLight,
              opacity: 0.04,
            ),
          ),
        ),
        Form(
          key: _formKey,
          child: ListView(
            padding: const EdgeInsets.all(AppSpacing.md),
            children: [
              // ── Basic info ────────────────────────────────────────────
              _Section(
                label: 'Жалпы ақпарат',
                child: Column(
                  children: [
                    TextFormField(
                      controller: _nameCtr,
                      textCapitalization: TextCapitalization.words,
                      decoration: const InputDecoration(
                        hintText: 'Аты-жөні',
                        prefixIcon: Icon(Icons.person_outline,
                            color: AppColors.textHint),
                      ),
                      validator: (v) =>
                          (v == null || v.trim().isEmpty)
                              ? 'Аты-жөніңізді енгізіңіз'
                              : null,
                    ),
                    const SizedBox(height: AppSpacing.md),
                    TextFormField(
                      controller: _ageCtr,
                      keyboardType: TextInputType.number,
                      decoration: const InputDecoration(
                        hintText: 'Жас (18–80)',
                        prefixIcon: Icon(Icons.cake_outlined,
                            color: AppColors.textHint),
                      ),
                      validator: (v) {
                        if (v == null || v.trim().isEmpty) return null;
                        final n = int.tryParse(v.trim());
                        if (n == null || n < 18 || n > 80) {
                          return '18–80 аралығын енгізіңіз';
                        }
                        return null;
                      },
                    ),
                    const SizedBox(height: AppSpacing.md),
                    TextFormField(
                      controller: _cityCtr,
                      textCapitalization: TextCapitalization.words,
                      decoration: const InputDecoration(
                        hintText: 'Қала',
                        prefixIcon: Icon(Icons.location_city_outlined,
                            color: AppColors.textHint),
                      ),
                    ),
                    const SizedBox(height: AppSpacing.md),
                    TextFormField(
                      controller: _bioCtr,
                      maxLines: 4,
                      maxLength: 500,
                      decoration: const InputDecoration(
                        hintText: 'Өзіңіз туралы...',
                        prefixIcon: Icon(Icons.edit_outlined,
                            color: AppColors.textHint),
                        alignLabelWithHint: true,
                      ),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: AppSpacing.md),

              // ── Niyyah ────────────────────────────────────────────────
              _Section(
                label: 'Ниет (мақсат)',
                child: Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: _niyyahOptions.map((opt) {
                    final (value, emoji, label) = opt;
                    final selected = _niyyah == value;
                    return GestureDetector(
                      onTap: () => setState(() => _niyyah = value),
                      child: AnimatedContainer(
                        duration: const Duration(milliseconds: 200),
                        padding: const EdgeInsets.symmetric(
                            horizontal: 16, vertical: 10),
                        decoration: BoxDecoration(
                          color: selected
                              ? AppColors.primary
                              : AppColors.surfaceVariant,
                          borderRadius: AppRadius.chip,
                          border: selected
                              ? Border.all(
                                  color: AppColors.goldBorder, width: 1.5)
                              : null,
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Text(emoji,
                                style: const TextStyle(fontSize: 16)),
                            const SizedBox(width: 6),
                            Text(
                              label,
                              style: GoogleFonts.nunito(
                                fontSize: 14,
                                fontWeight: FontWeight.w600,
                                color: selected
                                    ? Colors.white
                                    : AppColors.textSecondary,
                              ),
                            ),
                          ],
                        ),
                      ),
                    );
                  }).toList(),
                ),
              ),

              const SizedBox(height: AppSpacing.md),

              // ── Madhab ────────────────────────────────────────────────
              _Section(
                label: 'Мазхаб',
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    DropdownButtonFormField<String>(
                      value: _madhabOptions.any((o) => o.$1 == _madhab)
                          ? _madhab
                          : null,
                      decoration: const InputDecoration(
                        hintText: 'Мазхабты таңдаңыз',
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
                    if (_madhab != null) ...[
                      const SizedBox(height: 8),
                      MadhabBadge(madhab: _madhab!),
                    ],
                  ],
                ),
              ),

              const SizedBox(height: AppSpacing.md),

              // ── Languages ─────────────────────────────────────────────
              _Section(
                label: 'Тілдер',
                child: Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: _languageOptions.map((opt) {
                    final (value, label) = opt;
                    final selected = _languages.contains(value);
                    return GestureDetector(
                      onTap: () => setState(() {
                        if (selected) {
                          _languages.remove(value);
                        } else {
                          _languages.add(value);
                        }
                      }),
                      child: AnimatedContainer(
                        duration: const Duration(milliseconds: 200),
                        padding: const EdgeInsets.symmetric(
                            horizontal: 16, vertical: 10),
                        decoration: BoxDecoration(
                          color: selected
                              ? AppColors.secondary.withValues(alpha: 0.2)
                              : AppColors.surfaceVariant,
                          borderRadius: AppRadius.chip,
                          border: Border.all(
                            color: selected
                                ? AppColors.secondary
                                : Colors.transparent,
                            width: 1.5,
                          ),
                        ),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            if (selected) ...[
                              const Icon(Icons.check,
                                  size: 14, color: AppColors.secondary),
                              const SizedBox(width: 4),
                            ],
                            Text(
                              label,
                              style: GoogleFonts.nunito(
                                fontSize: 14,
                                fontWeight: FontWeight.w600,
                                color: selected
                                    ? AppColors.textPrimary
                                    : AppColors.textSecondary,
                              ),
                            ),
                          ],
                        ),
                      ),
                    );
                  }).toList(),
                ),
              ),

              const SizedBox(height: AppSpacing.md),

              // ── No-Photo Mode ─────────────────────────────────────────
              _Section(
                label: 'Құпиялылық',
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            'Фотосыз режим',
                            style: GoogleFonts.nunito(
                              fontSize: 15,
                              fontWeight: FontWeight.w600,
                              color: AppColors.textPrimary,
                            ),
                          ),
                          Text(
                            'Фото тек өзара лайктан кейін ашылады',
                            style: GoogleFonts.nunito(
                                fontSize: 12,
                                color: AppColors.textSecondary,
                                height: 1.4),
                          ),
                        ],
                      ),
                    ),
                    Switch(
                      value: _noPhotoMode,
                      activeColor: AppColors.primary,
                      onChanged: (v) => setState(() => _noPhotoMode = v),
                    ),
                  ],
                ),
              ),

              const SizedBox(height: AppSpacing.lg),

              // ── Error ─────────────────────────────────────────────────
              if (_error != null)
                Container(
                  padding: const EdgeInsets.all(AppSpacing.md),
                  margin: const EdgeInsets.only(bottom: AppSpacing.md),
                  decoration: BoxDecoration(
                    color: AppColors.accent.withValues(alpha: 0.1),
                    borderRadius: AppRadius.card,
                    border: Border.all(
                        color: AppColors.accent.withValues(alpha: 0.3)),
                  ),
                  child: Text(
                    _error!,
                    style: GoogleFonts.nunito(
                        fontSize: 13, color: AppColors.accent),
                  ),
                ),

              // ── Save ──────────────────────────────────────────────────
              SizedBox(
                width: double.infinity,
                child: ElevatedButton(
                  onPressed: _saving ? null : _save,
                  child: _saving
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(
                            color: Colors.white,
                            strokeWidth: 2,
                          ),
                        )
                      : const Text('Сақтау'),
                ),
              ),

              const SizedBox(height: AppSpacing.xxl),
            ],
          ),
        ),
      ],
    );
  }
}

// ─── Section card ─────────────────────────────────────────────────────────────

class _Section extends StatelessWidget {
  final String label;
  final Widget child;
  const _Section({required this.label, required this.child});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(AppSpacing.md),
      decoration: const BoxDecoration(
        color: AppColors.surface,
        borderRadius: AppRadius.card,
        boxShadow: AppShadows.soft,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const IslamicStarWidget(size: 13, color: AppColors.secondary),
              const SizedBox(width: 6),
              Text(
                label,
                style: GoogleFonts.nunito(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: AppColors.textSecondary,
                ),
              ),
            ],
          ),
          const SizedBox(height: AppSpacing.sm),
          child,
        ],
      ),
    );
  }
}
