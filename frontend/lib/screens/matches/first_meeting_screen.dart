import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';
import 'package:dio/dio.dart';
import '../../core/constants/api_constants.dart';
import '../../core/network/dio_error_message.dart';
import '../../core/theme/app_colors.dart';
import '../../core/theme/app_theme.dart';
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
  const FirstMeetingScreen({super.key, required this.matchId});

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
