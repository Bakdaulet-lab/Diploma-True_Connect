import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../core/constants/api_constants.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../models/settings.dart';
import '../../providers/auth_provider.dart';
import '../../providers/settings_provider.dart';
import '../../widgets/halal_pattern_painter.dart';

class SettingsScreen extends ConsumerWidget {
  const SettingsScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final asyncSettings = ref.watch(settingsNotifierProvider);

    return Scaffold(
      backgroundColor: AppColors.background,
      appBar: AppBar(
        backgroundColor: AppColors.primaryDark,
        centerTitle: true,
        bottom: const PreferredSize(
          preferredSize: Size.fromHeight(1),
          child:
              Divider(height: 1, thickness: 1, color: AppColors.goldBorder),
        ),
        title: Text(
          'Баптаулар',
          style: GoogleFonts.nunito(
            fontSize: 20,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
        ),
      ),
      body: asyncSettings.when(
        loading: () => const Center(
            child: CircularProgressIndicator(color: AppColors.primary)),
        error: (e, _) => _buildError(context, ref, e.toString()),
        data: (settings) => _buildSettings(context, ref, settings),
      ),
    );
  }

  Widget _buildSettings(
      BuildContext context, WidgetRef ref, UserSettings settings) {
    return ListView(
      padding: const EdgeInsets.all(AppSpacing.md),
      children: [
        // ── Halal Mode ────────────────────────────────────────────────────
        const _SectionHeader('Халал режим'),

        _SettingsCard(
          child: Column(
            children: [
              // Niyyah filter
              _DropdownRow(
                label: 'Ниет сүзгісі',
                subtitle: 'Тек осы ниетпен адамдарды көрсету',
                value: settings.niyyahFilter,
                items: const [
                  (null, 'Барлығы'),
                  ('nikah_year', 'Никах 🌙'),
                  ('serious_marriage', 'Маңызды'),
                  ('friendship', 'Достық'),
                ],
                onChanged: (v) => ref
                    .read(settingsNotifierProvider.notifier)
                    .update(settings.copyWith(niyyahFilter: v)),
              ),

              const Divider(height: 1, color: AppColors.divider),

              // Madhab filter
              _DropdownRow(
                label: 'Мазхаб сүзгісі',
                subtitle: 'Тек осы мазхабтан адамдарды көрсету',
                value: settings.madhabFilter,
                items: const [
                  (null, 'Барлығы'),
                  ('hanafi', 'Ханафи'),
                  ('shafii', 'Шафии'),
                  ('maliki', 'Маликий'),
                  ('hanbali', 'Ханбали'),
                ],
                onChanged: (v) => ref
                    .read(settingsNotifierProvider.notifier)
                    .update(settings.copyWith(madhabFilter: v)),
              ),

              const Divider(height: 1, color: AppColors.divider),

              // Modesty level slider
              Padding(
                padding: const EdgeInsets.symmetric(
                    horizontal: AppSpacing.md, vertical: AppSpacing.sm),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          'Ұстамдылық деңгейі',
                          style: GoogleFonts.nunito(
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                            color: AppColors.textPrimary,
                          ),
                        ),
                        Container(
                          padding: const EdgeInsets.symmetric(
                              horizontal: 10, vertical: 2),
                          decoration: const BoxDecoration(
                            color: AppColors.primaryLight,
                            borderRadius: AppRadius.chip,
                          ),
                          child: Text(
                            _modestyLabel(settings.modestyLevel),
                            style: GoogleFonts.nunito(
                              fontSize: 12,
                              color: AppColors.primary,
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                        ),
                      ],
                    ),
                    Slider(
                      value: settings.modestyLevel.toDouble(),
                      min: 0,
                      max: 3,
                      divisions: 3,
                      activeColor: AppColors.primary,
                      inactiveColor: AppColors.surfaceVariant,
                      onChanged: (v) => ref
                          .read(settingsNotifierProvider.notifier)
                          .update(settings.copyWith(
                              modestyLevel: v.toInt())),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),

        // ── Matching ──────────────────────────────────────────────────────
        const _SectionHeader('Іздеу'),

        _SettingsCard(
          child: Column(
            children: [
              _RangeRow(
                label: 'Жас диапазоны',
                min: settings.minAge.toDouble(),
                max: settings.maxAge.toDouble(),
                absMin: 18,
                absMax: 60,
                onChanged: (min, max) => ref
                    .read(settingsNotifierProvider.notifier)
                    .update(settings.copyWith(
                        minAge: min.toInt(), maxAge: max.toInt())),
              ),
              const Divider(height: 1, color: AppColors.divider),
              _SliderRow(
                label: 'Максималды қашықтық',
                value: settings.maxDistanceKm.toDouble(),
                min: 5,
                max: 200,
                unit: 'км',
                onChanged: (v) => ref
                    .read(settingsNotifierProvider.notifier)
                    .update(settings.copyWith(maxDistanceKm: v.toInt())),
              ),
            ],
          ),
        ),

        // ── Mahram Management ─────────────────────────────────────────────
        const _SectionHeader('Махрам басқару'),
        _MahramSection(ref: ref),

        // ── Account ───────────────────────────────────────────────────────
        const _SectionHeader('Аккаунт'),

        _SettingsCard(
          child: Column(
            children: [
              ListTile(
                leading: const Icon(Icons.verified_user_outlined,
                    color: AppColors.primary),
                title: Text('KYC верификация',
                    style: GoogleFonts.nunito(
                        fontSize: 14, color: AppColors.textPrimary)),
                trailing: const Icon(Icons.chevron_right,
                    color: AppColors.textHint),
                onTap: () => context.push('/kyc'),
              ),
              const Divider(height: 1, color: AppColors.divider),
              ListTile(
                leading:
                    const Icon(Icons.logout, color: AppColors.accent),
                title: Text('Шығу',
                    style: GoogleFonts.nunito(
                        fontSize: 14, color: AppColors.accent)),
                onTap: () async {
                  await ref.read(authStateProvider.notifier).logout();
                  if (context.mounted) context.go('/auth/login');
                },
              ),
            ],
          ),
        ),

        const SizedBox(height: AppSpacing.xxl),
      ],
    );
  }

  Widget _buildError(BuildContext context, WidgetRef ref, String msg) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(Icons.error_outline, size: 48, color: AppColors.textHint),
          const SizedBox(height: 16),
          Text(msg,
              style: GoogleFonts.nunito(color: AppColors.textSecondary),
              textAlign: TextAlign.center),
          const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () => ref.refresh(settingsNotifierProvider),
            child: const Text('Қайтадан'),
          ),
        ],
      ),
    );
  }

  String _modestyLabel(int level) {
    switch (level) {
      case 1:
        return 'Орташа';
      case 2:
        return 'Жоғары';
      case 3:
        return 'Ең жоғары';
      default:
        return 'Стандарт';
    }
  }
}

// ─── Section Header ───────────────────────────────────────────────────────────

class _SectionHeader extends StatelessWidget {
  final String text;
  const _SectionHeader(this.text);

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(4, 20, 4, 8),
      child: Row(
        children: [
          const IslamicStarWidget(size: 12, color: AppColors.secondary),
          const SizedBox(width: 6),
          Text(
            text.toUpperCase(),
            style: GoogleFonts.nunito(
              fontSize: 11,
              fontWeight: FontWeight.bold,
              color: AppColors.textHint,
              letterSpacing: 1.2,
            ),
          ),
        ],
      ),
    );
  }
}

// ─── Settings Card ────────────────────────────────────────────────────────────

class _SettingsCard extends StatelessWidget {
  final Widget child;
  const _SettingsCard({required this.child});

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: const BoxDecoration(
        color: AppColors.surface,
        borderRadius: AppRadius.card,
        boxShadow: AppShadows.soft,
      ),
      clipBehavior: Clip.antiAlias,
      child: child,
    );
  }
}

// ─── Dropdown Row ─────────────────────────────────────────────────────────────

class _DropdownRow extends StatelessWidget {
  final String label;
  final String subtitle;
  final String? value;
  final List<(String?, String)> items;
  final ValueChanged<String?> onChanged;

  const _DropdownRow({
    required this.label,
    required this.subtitle,
    required this.value,
    required this.items,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(
          horizontal: AppSpacing.md, vertical: AppSpacing.sm),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(label,
                    style: GoogleFonts.nunito(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: AppColors.textPrimary,
                    )),
                Text(subtitle,
                    style: GoogleFonts.nunito(
                        fontSize: 11, color: AppColors.textHint)),
              ],
            ),
          ),
          DropdownButton<String?>(
            value: value,
            underline: const SizedBox.shrink(),
            style: GoogleFonts.nunito(
                fontSize: 13, color: AppColors.textPrimary),
            items: items
                .map((pair) => DropdownMenuItem<String?>(
                      value: pair.$1,
                      child: Text(pair.$2),
                    ))
                .toList(),
            onChanged: onChanged,
          ),
        ],
      ),
    );
  }
}

// ─── Range Row ────────────────────────────────────────────────────────────────

class _RangeRow extends StatelessWidget {
  final String label;
  final double min;
  final double max;
  final double absMin;
  final double absMax;
  final void Function(double min, double max) onChanged;

  const _RangeRow({
    required this.label,
    required this.min,
    required this.max,
    required this.absMin,
    required this.absMax,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(
          horizontal: AppSpacing.md, vertical: AppSpacing.sm),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(label,
                  style: GoogleFonts.nunito(
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                    color: AppColors.textPrimary,
                  )),
              Text('${min.toInt()} – ${max.toInt()} жас',
                  style: GoogleFonts.nunito(
                      fontSize: 12, color: AppColors.textSecondary)),
            ],
          ),
          RangeSlider(
            values: RangeValues(min, max),
            min: absMin,
            max: absMax,
            divisions: (absMax - absMin).toInt(),
            activeColor: AppColors.primary,
            inactiveColor: AppColors.surfaceVariant,
            onChanged: (v) => onChanged(v.start, v.end),
          ),
        ],
      ),
    );
  }
}

// ─── Slider Row ───────────────────────────────────────────────────────────────

class _SliderRow extends StatelessWidget {
  final String label;
  final double value;
  final double min;
  final double max;
  final String unit;
  final ValueChanged<double> onChanged;

  const _SliderRow({
    required this.label,
    required this.value,
    required this.min,
    required this.max,
    required this.unit,
    required this.onChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(
          horizontal: AppSpacing.md, vertical: AppSpacing.sm),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(label,
                  style: GoogleFonts.nunito(
                    fontSize: 14,
                    fontWeight: FontWeight.w600,
                    color: AppColors.textPrimary,
                  )),
              Text('${value.toInt()} $unit',
                  style: GoogleFonts.nunito(
                      fontSize: 12, color: AppColors.textSecondary)),
            ],
          ),
          Slider(
            value: value,
            min: min,
            max: max,
            activeColor: AppColors.primary,
            inactiveColor: AppColors.surfaceVariant,
            onChanged: onChanged,
          ),
        ],
      ),
    );
  }
}

// ─── Mahram Section ───────────────────────────────────────────────────────────

class _MahramSection extends StatefulWidget {
  final WidgetRef ref;
  const _MahramSection({required this.ref});

  @override
  State<_MahramSection> createState() => _MahramSectionState();
}

class _MahramSectionState extends State<_MahramSection> {
  final _phoneCtr = TextEditingController();
  bool _showAdd = false;
  bool _loading = false;
  List<Map<String, dynamic>> _mahrams = [];

  @override
  void initState() {
    super.initState();
    _loadMahrams();
  }

  Future<void> _loadMahrams() async {
    try {
      final dio = widget.ref.read(dioClientProvider).dio;
      final resp = await dio.get(ApiConstants.mahram);
      if (!mounted) return;
      setState(() {
        _mahrams = (resp.data as List<dynamic>)
            .map((e) => e as Map<String, dynamic>)
            .toList();
      });
    } catch (_) {}
  }

  Future<void> _addMahram() async {
    if (_phoneCtr.text.trim().isEmpty) return;
    setState(() => _loading = true);
    try {
      final dio = widget.ref.read(dioClientProvider).dio;
      await dio.post(ApiConstants.mahram,
          data: {'phone': _phoneCtr.text.trim()});
      _phoneCtr.clear();
      if (!mounted) return;
      setState(() {
        _showAdd = false;
        _loading = false;
      });
      await _loadMahrams();
    } catch (_) {
      if (!mounted) return;
      setState(() => _loading = false);
    }
  }

  @override
  void dispose() {
    _phoneCtr.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return _SettingsCard(
      child: Column(
        children: [
          // Existing mahrams
          ..._mahrams.map((m) => _MahramTile(mahram: m)),

          // Add button
          if (!_showAdd)
            ListTile(
              leading: const Icon(Icons.add_circle_outline,
                  color: AppColors.primary),
              title: Text(
                'Махрам қосу',
                style: GoogleFonts.nunito(
                    fontSize: 14, color: AppColors.primary),
              ),
              onTap: () => setState(() => _showAdd = true),
            )
          else
            Padding(
              padding: const EdgeInsets.all(AppSpacing.md),
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      controller: _phoneCtr,
                      keyboardType: TextInputType.phone,
                      decoration: const InputDecoration(
                        hintText: '+7 ___ ___ __ __',
                        prefixIcon: Icon(Icons.phone_outlined,
                            color: AppColors.textHint),
                      ),
                    ),
                  ),
                  const SizedBox(width: 8),
                  ElevatedButton(
                    onPressed: _loading ? null : _addMahram,
                    style: ElevatedButton.styleFrom(
                        padding: const EdgeInsets.symmetric(
                            horizontal: AppSpacing.md)),
                    child: _loading
                        ? const SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(
                                color: Colors.white, strokeWidth: 2))
                        : const Text('Қосу'),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }
}

class _MahramTile extends StatelessWidget {
  final Map<String, dynamic> mahram;
  const _MahramTile({required this.mahram});

  @override
  Widget build(BuildContext context) {
    final status = mahram['verification_status'] as String? ?? 'pending';
    final isVerified = status == 'verified';

    return ListTile(
      leading: CircleAvatar(
        backgroundColor:
            isVerified ? AppColors.primaryLight : AppColors.surfaceVariant,
        child: Icon(
          isVerified ? Icons.check : Icons.hourglass_top,
          color: isVerified ? AppColors.primary : AppColors.textHint,
          size: 18,
        ),
      ),
      title: Text(
        'Махрам',
        style: GoogleFonts.nunito(
            fontSize: 14, color: AppColors.textPrimary),
      ),
      subtitle: Text(
        isVerified ? 'Расталған' : 'Күтілуде',
        style: GoogleFonts.nunito(
          fontSize: 12,
          color: isVerified ? AppColors.primary : AppColors.textHint,
        ),
      ),
    );
  }
}
