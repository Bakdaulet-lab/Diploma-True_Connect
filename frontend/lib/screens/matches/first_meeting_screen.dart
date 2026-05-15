import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:dio/dio.dart';
import '../../core/constants/api_constants.dart';
import '../../core/network/dio_error_message.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
import '../../models/interaction.dart';
import '../../models/venue.dart';
import '../../providers/auth_provider.dart';
import '../../widgets/halal_pattern_painter.dart';

// ─── Provider ─────────────────────────────────────────────────────────────────

final _venuesProvider =
    FutureProvider.family<List<Venue>, String>((ref, city) async {
  final dio = ref.watch(dioClientProvider).dio;
  try {
    final resp = await dio.get(
      ApiConstants.venues,
      queryParameters: city.isNotEmpty ? {'city': city} : null,
    );
    final raw = resp.data;
    final list = raw is Map ? raw['data'] ?? raw : raw;
    if (list is List) {
      return list
          .whereType<Map>()
          .map((e) => Venue.fromJson(Map<String, dynamic>.from(e)))
          .toList();
    }
    return <Venue>[];
  } on DioException catch (e) {
    throw dioErrorMessage(e);
  }
});

// ─── Screen ───────────────────────────────────────────────────────────────────

class FirstMeetingScreen extends ConsumerStatefulWidget {
  final String matchId;
  final String otherUserId;
  const FirstMeetingScreen({
    super.key,
    required this.matchId,
    this.otherUserId = '',
  });

  @override
  ConsumerState<FirstMeetingScreen> createState() => _FirstMeetingScreenState();
}

class _FirstMeetingScreenState extends ConsumerState<FirstMeetingScreen> {
  static const _cities = [
    'Almaty',
    'Astana',
    'Shymkent',
    'Karagandy',
    'Aktobe',
  ];

  static const _cityLabels = {
    'Almaty': 'Алматы',
    'Astana': 'Астана',
    'Shymkent': 'Шымкент',
    'Karagandy': 'Қарағанды',
    'Aktobe': 'Ақтөбе',
  };

  String _selectedCity = 'Almaty';

  void _showRateSheet(BuildContext context) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: AppColors.surface,
      shape: const RoundedRectangleBorder(borderRadius: AppRadius.bottomSheet),
      builder: (_) => _RateMeetingSheet(
        matchId: widget.matchId,
        otherUserId: widget.otherUserId,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final asyncVenues = ref.watch(_venuesProvider(_selectedCity));

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
          child:
              Divider(height: 1, thickness: 1, color: AppColors.goldBorder),
        ),
        title: Text(
          'Бірінші кездесу',
          style: GoogleFonts.nunito(
            fontSize: 20,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
        ),
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showRateSheet(context),
        backgroundColor: AppColors.secondary,
        foregroundColor: AppColors.textPrimary,
        icon: const Icon(Icons.star_outline),
        label: Text(
          'Кездесуді бағала',
          style: GoogleFonts.nunito(fontWeight: FontWeight.w600),
        ),
      ),
      body: Stack(
        children: [
          const Positioned.fill(
            child: CustomPaint(
              painter: HalalPatternPainter(
                color: AppColors.primaryLight,
                opacity: 0.04,
              ),
            ),
          ),
          CustomScrollView(
            slivers: [
              // ── Halal meeting rules ───────────────────────────────────
              SliverToBoxAdapter(
                child: Container(
                  margin: const EdgeInsets.all(AppSpacing.md),
                  padding: const EdgeInsets.all(AppSpacing.md),
                  decoration: BoxDecoration(
                    color: AppColors.primary.withValues(alpha: 0.1),
                    borderRadius: AppRadius.card,
                    border: Border.all(
                        color: AppColors.goldBorder.withValues(alpha: 0.5)),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          const IslamicStarWidget(
                              size: 16, color: AppColors.secondary),
                          const SizedBox(width: 8),
                          Text(
                            'Халяль кездесу ережелері',
                            style: GoogleFonts.nunito(
                              fontSize: 14,
                              fontWeight: FontWeight.bold,
                              color: AppColors.textPrimary,
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: AppSpacing.sm),
                      const _Rule(
                        num: '1',
                        text:
                            'Қоғамдық жерде кездесіңіз — кафе, саябақ немесе мәдени орын.',
                      ),
                      const _Rule(
                        num: '2',
                        text:
                            'Махрамды шақыруды ұмытпаңыз немесе ашық кеңістікті таңдаңыз.',
                      ),
                      const _Rule(
                        num: '3',
                        text:
                            'Кездесу мақсаты — танысу. Асықпаңыз, Аллаһ береке берсін.',
                      ),
                    ],
                  ),
                ),
              ),

              // ── City selector ─────────────────────────────────────────
              SliverToBoxAdapter(
                child: Padding(
                  padding: const EdgeInsets.symmetric(
                      horizontal: AppSpacing.md),
                  child: SingleChildScrollView(
                    scrollDirection: Axis.horizontal,
                    child: Row(
                      children: _cities.map((city) {
                        final selected = _selectedCity == city;
                        return GestureDetector(
                          onTap: () =>
                              setState(() => _selectedCity = city),
                          child: AnimatedContainer(
                            duration: const Duration(milliseconds: 200),
                            margin: const EdgeInsets.only(right: 8),
                            padding: const EdgeInsets.symmetric(
                                horizontal: 16, vertical: 8),
                            decoration: BoxDecoration(
                              color: selected
                                  ? AppColors.primary
                                  : AppColors.surface,
                              borderRadius: AppRadius.chip,
                              border: Border.all(
                                color: selected
                                    ? AppColors.goldBorder
                                    : AppColors.surfaceVariant,
                              ),
                            ),
                            child: Text(
                              _cityLabels[city] ?? city,
                              style: GoogleFonts.nunito(
                                fontSize: 13,
                                fontWeight: FontWeight.w600,
                                color: selected
                                    ? Colors.white
                                    : AppColors.textSecondary,
                              ),
                            ),
                          ),
                        );
                      }).toList(),
                    ),
                  ),
                ),
              ),

              const SliverToBoxAdapter(
                  child: SizedBox(height: AppSpacing.md)),

              // ── Venue list ────────────────────────────────────────────
              asyncVenues.when(
                loading: () => const SliverToBoxAdapter(
                  child: Center(
                    child: Padding(
                      padding: EdgeInsets.all(32),
                      child: CircularProgressIndicator(
                          color: AppColors.primary),
                    ),
                  ),
                ),
                error: (e, _) => SliverToBoxAdapter(
                  child: Center(
                    child: Padding(
                      padding: const EdgeInsets.all(AppSpacing.lg),
                      child: Text(e.toString(),
                          style: GoogleFonts.nunito(
                              color: AppColors.textSecondary)),
                    ),
                  ),
                ),
                data: (venueList) => venueList.isEmpty
                    ? SliverToBoxAdapter(
                        child: Center(
                          child: Padding(
                            padding:
                                const EdgeInsets.all(AppSpacing.xl),
                            child: Column(
                              children: [
                                const IslamicStarWidget(
                                    size: 48,
                                    color: AppColors.surfaceVariant),
                                const SizedBox(height: 16),
                                Text(
                                  'Бұл қалада орындар жоқ',
                                  style: GoogleFonts.nunito(
                                      color: AppColors.textSecondary),
                                ),
                              ],
                            ),
                          ),
                        ),
                      )
                    : SliverPadding(
                        padding: const EdgeInsets.symmetric(
                            horizontal: AppSpacing.md),
                        sliver: SliverList(
                          delegate: SliverChildBuilderDelegate(
                            (ctx, i) =>
                                _VenueCard(venue: venueList[i]),
                            childCount: venueList.length,
                          ),
                        ),
                      ),
              ),

              const SliverToBoxAdapter(
                  child: SizedBox(height: AppSpacing.xxl)),
            ],
          ),
        ],
      ),
    );
  }
}

// ─── Rule row ─────────────────────────────────────────────────────────────────

class _Rule extends StatelessWidget {
  final String num;
  final String text;
  const _Rule({required this.num, required this.text});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: 6),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: 20,
            height: 20,
            alignment: Alignment.center,
            decoration: const BoxDecoration(
              color: AppColors.secondary,
              shape: BoxShape.circle,
            ),
            child: Text(
              num,
              style: GoogleFonts.nunito(
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary),
            ),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              text,
              style: GoogleFonts.nunito(
                  fontSize: 13,
                  color: AppColors.textSecondary,
                  height: 1.5),
            ),
          ),
        ],
      ),
    );
  }
}

// ─── Venue card ───────────────────────────────────────────────────────────────

class _VenueCard extends StatelessWidget {
  final Venue venue;
  const _VenueCard({required this.venue});

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: AppSpacing.sm),
      padding: const EdgeInsets.all(AppSpacing.md),
      decoration: const BoxDecoration(
        color: AppColors.surface,
        borderRadius: AppRadius.card,
        boxShadow: AppShadows.soft,
      ),
      child: Row(
        children: [
          // Category icon
          Container(
            width: 48,
            height: 48,
            alignment: Alignment.center,
            decoration: BoxDecoration(
              color: AppColors.primaryLight.withValues(alpha: 0.15),
              borderRadius: AppRadius.chip,
            ),
            child: Text(
              venue.categoryEmoji,
              style: const TextStyle(fontSize: 22),
            ),
          ),

          const SizedBox(width: AppSpacing.md),

          // Info
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  venue.name,
                  style: GoogleFonts.nunito(
                    fontSize: 15,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
                const SizedBox(height: 2),
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 6, vertical: 2),
                      decoration: BoxDecoration(
                        color: AppColors.secondary.withValues(alpha: 0.15),
                        borderRadius: AppRadius.chip,
                      ),
                      child: Text(
                        venue.categoryLabel,
                        style: GoogleFonts.nunito(
                            fontSize: 10,
                            fontWeight: FontWeight.w600,
                            color: AppColors.textSecondary),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 4),
                Row(
                  children: [
                    const Icon(Icons.location_on_outlined,
                        size: 12, color: AppColors.textHint),
                    const SizedBox(width: 3),
                    Expanded(
                      child: Text(
                        venue.address,
                        style: GoogleFonts.nunito(
                            fontSize: 12, color: AppColors.textHint),
                        maxLines: 1,
                        overflow: TextOverflow.ellipsis,
                      ),
                    ),
                  ],
                ),
                if (venue.phone != null && venue.phone!.isNotEmpty) ...[
                  const SizedBox(height: 2),
                  Row(
                    children: [
                      const Icon(Icons.phone_outlined,
                          size: 12, color: AppColors.textHint),
                      const SizedBox(width: 3),
                      Text(
                        venue.phone!,
                        style: GoogleFonts.nunito(
                            fontSize: 12, color: AppColors.textHint),
                      ),
                    ],
                  ),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }
}

// ─── Rate meeting sheet ───────────────────────────────────────────────────────

class _RateMeetingSheet extends ConsumerStatefulWidget {
  final String matchId;
  final String otherUserId;
  const _RateMeetingSheet({required this.matchId, required this.otherUserId});

  @override
  ConsumerState<_RateMeetingSheet> createState() => _RateMeetingSheetState();
}

class _RateMeetingSheetState extends ConsumerState<_RateMeetingSheet> {
  int _rating = 3;
  InteractionContext _ctx = InteractionContext.date;
  final _commentCtrl = TextEditingController();
  bool _loading = false;
  bool _done = false;

  @override
  void dispose() {
    _commentCtrl.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    setState(() => _loading = true);
    try {
      final dio = ref.read(dioClientProvider).dio;
      if (widget.otherUserId.isEmpty) {
        throw Exception('Пайдаланушы ID анықталмады');
      }
      await dio.post(
        ApiConstants.interactions,
        data: {
          'rated_id': widget.otherUserId,
          'rating': _rating,
          'context': _ctx.value,
          if (_commentCtrl.text.trim().isNotEmpty)
            'comment': _commentCtrl.text.trim(),
        },
      );
      setState(() => _done = true);
    } on DioException catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(dioErrorMessage(e))),
        );
      }
    } finally {
      if (mounted) setState(() => _loading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final bottom = MediaQuery.of(context).viewInsets.bottom;
    return Padding(
      padding: EdgeInsets.only(bottom: bottom),
      child: Container(
        padding: const EdgeInsets.all(AppSpacing.lg),
        child: _done ? _buildSuccess() : _buildForm(),
      ),
    );
  }

  Widget _buildSuccess() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        const Icon(Icons.check_circle, color: AppColors.primary, size: 56),
        const SizedBox(height: AppSpacing.md),
        Text(
          'Бағалау жіберілді!',
          style: GoogleFonts.nunito(
            fontSize: 18, fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
          ),
        ),
        const SizedBox(height: AppSpacing.sm),
        Text(
          'Trust Score жаңартылады.',
          style: GoogleFonts.nunito(fontSize: 13, color: AppColors.textSecondary),
        ),
        const SizedBox(height: AppSpacing.lg),
        ElevatedButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Жабу'),
        ),
      ],
    );
  }

  Widget _buildForm() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Center(
          child: Container(
            width: 40, height: 4,
            decoration: BoxDecoration(
              color: AppColors.divider,
              borderRadius: BorderRadius.circular(2),
            ),
          ),
        ),
        const SizedBox(height: AppSpacing.md),
        Text(
          'Кездесуді бағала',
          style: GoogleFonts.nunito(
            fontSize: 18, fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
          ),
        ),
        const SizedBox(height: AppSpacing.md),
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: List.generate(5, (i) {
            final star = i + 1;
            return IconButton(
              onPressed: () => setState(() => _rating = star),
              icon: Icon(
                star <= _rating ? Icons.star : Icons.star_border,
                color: AppColors.secondary,
                size: 36,
              ),
            );
          }),
        ),
        const SizedBox(height: AppSpacing.md),
        Text(
          'Кездесу түрі',
          style: GoogleFonts.nunito(
            fontSize: 13, fontWeight: FontWeight.w600,
            color: AppColors.textSecondary,
          ),
        ),
        const SizedBox(height: AppSpacing.xs),
        Wrap(
          spacing: AppSpacing.sm,
          children: InteractionContext.values.map((c) {
            final selected = _ctx == c;
            return ChoiceChip(
              label: Text(c.label),
              selected: selected,
              onSelected: (_) => setState(() => _ctx = c),
              selectedColor: AppColors.primary,
              labelStyle: GoogleFonts.nunito(
                fontSize: 13,
                color: selected ? Colors.white : AppColors.textSecondary,
              ),
            );
          }).toList(),
        ),
        const SizedBox(height: AppSpacing.md),
        TextField(
          controller: _commentCtrl,
          maxLines: 2,
          maxLength: 300,
          decoration: const InputDecoration(hintText: 'Пікір (міндетті емес)...'),
        ),
        const SizedBox(height: AppSpacing.md),
        ElevatedButton(
          onPressed: _loading ? null : _submit,
          child: _loading
              ? const SizedBox(
                  width: 20, height: 20,
                  child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                )
              : const Text('Жіберу'),
        ),
      ],
    );
  }
}
