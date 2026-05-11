import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import '../../core/constants/api_constants.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../providers/auth_provider.dart';
import '../../widgets/halal_pattern_painter.dart';

// Imam data model (mirrors internal/pkg/imam/catalog.go)
class _Imam {
  final String id;
  final String name;
  final String city;
  final String mosque;
  final List<String> languages;
  final String? availabilityNotes;

  const _Imam({
    required this.id,
    required this.name,
    required this.city,
    required this.mosque,
    required this.languages,
    this.availabilityNotes,
  });

  factory _Imam.fromJson(Map<String, dynamic> j) => _Imam(
        id: j['id'] as String,
        name: j['name'] as String,
        city: j['city'] as String,
        mosque: j['mosque'] as String,
        languages: (j['languages'] as List<dynamic>?)
                ?.map((e) => e as String)
                .toList() ??
            [],
        availabilityNotes: j['availability_notes'] as String?,
      );
}

final _imamsProvider =
    FutureProvider.family<List<_Imam>, String>((ref, city) async {
  final dio = ref.watch(dioClientProvider).dio;
  final resp = await dio.get(
    ApiConstants.imams,
    queryParameters: city.isNotEmpty ? {'city': city} : null,
  );
  return (resp.data as List<dynamic>)
      .map((e) => _Imam.fromJson(e as Map<String, dynamic>))
      .toList();
});

class ImamConnectScreen extends ConsumerStatefulWidget {
  final String? matchId;
  const ImamConnectScreen({super.key, this.matchId});

  @override
  ConsumerState<ImamConnectScreen> createState() =>
      _ImamConnectScreenState();
}

class _ImamConnectScreenState extends ConsumerState<ImamConnectScreen> {
  String _selectedCity = 'Алматы';

  static const _cities = [
    'Алматы',
    'Астана',
    'Шымкент',
    'Қарағанды',
    'Ақтөбе',
  ];

  @override
  Widget build(BuildContext context) {
    final asyncImams = ref.watch(_imamsProvider(_selectedCity));

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
        title: Column(
          children: [
            Text(
              'Имам Connect',
              style: GoogleFonts.nunito(
                fontSize: 20,
                fontWeight: FontWeight.bold,
                color: Colors.white,
              ),
            ),
            Text(
              'Никахты растаңыз',
              style: GoogleFonts.nunito(
                fontSize: 10,
                color: AppColors.secondary,
                letterSpacing: 1.5,
              ),
            ),
          ],
        ),
      ),
      body: Column(
        children: [
          // City selector
          Container(
            color: AppColors.surface,
            padding: const EdgeInsets.symmetric(
                horizontal: AppSpacing.lg, vertical: AppSpacing.sm),
            child: Row(
              children: [
                const Icon(Icons.location_city,
                    size: 18, color: AppColors.primary),
                const SizedBox(width: 8),
                Text('Қала:',
                    style: GoogleFonts.nunito(
                        fontSize: 13, color: AppColors.textSecondary)),
                const SizedBox(width: 8),
                DropdownButton<String>(
                  value: _selectedCity,
                  underline: const SizedBox.shrink(),
                  style: GoogleFonts.nunito(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: AppColors.textPrimary),
                  items: _cities
                      .map((c) => DropdownMenuItem(value: c, child: Text(c)))
                      .toList(),
                  onChanged: (v) {
                    if (v != null) setState(() => _selectedCity = v);
                  },
                ),
              ],
            ),
          ),

          Expanded(
            child: asyncImams.when(
              loading: () => const Center(
                  child:
                      CircularProgressIndicator(color: AppColors.primary)),
              error: (e, _) => _buildError(e.toString()),
              data: (imams) => imams.isEmpty
                  ? _buildEmpty()
                  : ListView.separated(
                      padding: const EdgeInsets.all(AppSpacing.md),
                      itemCount: imams.length,
                      separatorBuilder: (_, __) =>
                          const SizedBox(height: AppSpacing.sm),
                      itemBuilder: (_, i) => _ImamCard(
                        imam: imams[i],
                        matchId: widget.matchId,
                      ),
                    ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildEmpty() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const IslamicStarWidget(size: 48, color: AppColors.secondary),
          const SizedBox(height: 16),
          Text('Имам табылмады',
              style: GoogleFonts.nunito(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary)),
          const SizedBox(height: 8),
          Text('Басқа қаланы таңдаңыз',
              style: GoogleFonts.nunito(
                  fontSize: 13, color: AppColors.textSecondary)),
        ],
      ),
    );
  }

  Widget _buildError(String msg) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(AppSpacing.xl),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.wifi_off, size: 48, color: AppColors.textHint),
            const SizedBox(height: 16),
            Text(msg,
                textAlign: TextAlign.center,
                style:
                    GoogleFonts.nunito(color: AppColors.textSecondary)),
          ],
        ),
      ),
    );
  }
}

// ─── Imam Card ────────────────────────────────────────────────────────────────

class _ImamCard extends StatelessWidget {
  final _Imam imam;
  final String? matchId;

  const _ImamCard({required this.imam, required this.matchId});

  @override
  Widget build(BuildContext context) {
    return Container(
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
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Avatar placeholder with Islamic star
              Container(
                width: 52,
                height: 52,
                decoration: BoxDecoration(
                  color: AppColors.primaryLight,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: const Center(
                  child: IslamicStarWidget(
                      size: 28, color: AppColors.primary),
                ),
              ),
              const SizedBox(width: AppSpacing.md),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      imam.name,
                      style: GoogleFonts.nunito(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: AppColors.textPrimary,
                      ),
                    ),
                    const SizedBox(height: 2),
                    Text(
                      imam.mosque,
                      style: GoogleFonts.nunito(
                          fontSize: 13, color: AppColors.textSecondary),
                    ),
                    const SizedBox(height: 2),
                    Row(
                      children: [
                        const Icon(Icons.location_on_outlined,
                            size: 13, color: AppColors.textHint),
                        const SizedBox(width: 3),
                        Text(
                          imam.city,
                          style: GoogleFonts.nunito(
                              fontSize: 12, color: AppColors.textHint),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),

          if (imam.availabilityNotes != null) ...[
            const SizedBox(height: 8),
            Text(
              imam.availabilityNotes!,
              style: GoogleFonts.nunito(
                  fontSize: 12,
                  color: AppColors.textSecondary,
                  fontStyle: FontStyle.italic),
            ),
          ],

          const SizedBox(height: 8),

          // Language chips
          Wrap(
            spacing: 6,
            children: imam.languages
                .map((l) => Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 8, vertical: 3),
                      decoration: const BoxDecoration(
                        color: AppColors.surfaceVariant,
                        borderRadius: AppRadius.chip,
                      ),
                      child: Text(l,
                          style: GoogleFonts.nunito(
                              fontSize: 11,
                              color: AppColors.textSecondary)),
                    ))
                .toList(),
          ),

          if (matchId != null) ...[
            const SizedBox(height: AppSpacing.md),
            SizedBox(
              width: double.infinity,
              child: ElevatedButton(
                onPressed: () =>
                    _confirmNikah(context, imam),
                child: const Text('Никахты растау'),
              ),
            ),
          ],
        ],
      ),
    );
  }

  void _confirmNikah(BuildContext context, _Imam imam) {
    showDialog(
      context: context,
      builder: (_) => _NikahConfirmDialog(
        imam: imam,
        matchId: matchId!,
      ),
    );
  }
}

// ─── Nikah Confirm Dialog ─────────────────────────────────────────────────────

class _NikahConfirmDialog extends ConsumerStatefulWidget {
  final _Imam imam;
  final String matchId;

  const _NikahConfirmDialog({required this.imam, required this.matchId});

  @override
  ConsumerState<_NikahConfirmDialog> createState() =>
      _NikahConfirmDialogState();
}

class _NikahConfirmDialogState extends ConsumerState<_NikahConfirmDialog> {
  bool _loading = false;
  bool _confirmed = false;

  Future<void> _confirm() async {
    setState(() => _loading = true);
    try {
      final dio = ref.read(dioClientProvider).dio;
      await dio.post(
        '/matches/${widget.matchId}/nikah-confirm',
        data: {'imam_id': widget.imam.id},
      );
      if (mounted) setState(() => _confirmed = true);
    } catch (_) {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    if (_confirmed) {
      return AlertDialog(
        backgroundColor: AppColors.surface,
        shape: const RoundedRectangleBorder(borderRadius: AppRadius.card),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            const IslamicStarWidget(size: 48, color: AppColors.secondary),
            const SizedBox(height: 16),
            Text(
              'Никах расталды',
              style: GoogleFonts.nunito(
                fontSize: 20,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              'بَارَكَ اللَّهُ لَكُمَا',
              style: GoogleFonts.nunito(
                fontSize: 22,
                color: AppColors.secondary,
                fontWeight: FontWeight.bold,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 4),
            Text(
              'Барака Аллаху фикум 🌙',
              style: GoogleFonts.nunito(
                  fontSize: 14, color: AppColors.textSecondary),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 8),
            Text(
              'Аккаунттарыңыз белсенді, хат тарихы сақталды.',
              style: GoogleFonts.nunito(
                  fontSize: 12, color: AppColors.textHint),
              textAlign: TextAlign.center,
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              context.go('/home');
            },
            child: Text('Жабу',
                style: GoogleFonts.nunito(color: AppColors.primary)),
          ),
        ],
      );
    }

    return AlertDialog(
      backgroundColor: AppColors.surface,
      shape: const RoundedRectangleBorder(borderRadius: AppRadius.card),
      title: Text(
        'Никахты растау',
        style: GoogleFonts.nunito(
          fontSize: 18,
          fontWeight: FontWeight.bold,
          color: AppColors.textPrimary,
        ),
      ),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Имам: ${widget.imam.name}',
            style: GoogleFonts.nunito(
                fontSize: 14, color: AppColors.textPrimary),
          ),
          const SizedBox(height: 4),
          Text(
            widget.imam.mosque,
            style: GoogleFonts.nunito(
                fontSize: 13, color: AppColors.textSecondary),
          ),
          const SizedBox(height: 16),
          Container(
            padding: const EdgeInsets.all(AppSpacing.sm),
            decoration: const BoxDecoration(
              color: AppColors.secondaryLight,
              borderRadius: AppRadius.chip,
            ),
            child: Text(
              'Бұл іс-әрекетті болдырмау мүмкін емес. '
              'Растаудан кейін профильдер "никахталған" деп белгіленеді '
              'және іздеуден жасырылады.',
              style: GoogleFonts.nunito(
                  fontSize: 12,
                  color: AppColors.secondary,
                  height: 1.5),
            ),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: _loading ? null : () => Navigator.pop(context),
          child: Text('Бас тарту',
              style: GoogleFonts.nunito(color: AppColors.textSecondary)),
        ),
        ElevatedButton(
          onPressed: _loading ? null : _confirm,
          child: _loading
              ? const SizedBox(
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(
                      color: Colors.white, strokeWidth: 2))
              : const Text('Растау'),
        ),
      ],
    );
  }
}
