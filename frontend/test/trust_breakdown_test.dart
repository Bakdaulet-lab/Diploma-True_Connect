import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:trueconnect/models/trust_breakdown.dart';
import 'package:trueconnect/providers/trust_provider.dart';
import 'package:trueconnect/screens/profile/trust_breakdown_screen.dart';

void main() {
  group('TrustBreakdown.fromJson', () {
    test('parses the {breakdown, ratings} envelope', () {
      final b = TrustBreakdown.fromJson({
        'breakdown': {
          'score': 72,
          'badge': 'Gold / Member',
          'rating_count': 4,
          'smoothed_rating': 3.6,
          'base_score': 72.0,
          'kyc_bonus': 10.0,
          'report_count': 1,
          'report_penalty': 15.0,
          'raw_score': 67.0,
        },
        'ratings': [
          {'id': 'r1', 'rating': 5, 'context': 'date', 'comment': 'жақсы'},
          {'id': 'r2', 'rating': 4, 'context': 'meetup', 'comment': ''},
        ],
      });
      expect(b.score, 72);
      expect(b.badge, 'Gold / Member');
      expect(b.ratingCount, 4);
      expect(b.smoothedRating, closeTo(3.6, 1e-9));
      expect(b.kycBonus, 10.0);
      expect(b.reportPenalty, 15.0);
      expect(b.recentRatings.length, 2);
      expect(b.recentRatings.first.rating, 5);
      expect(b.recentRatings.first.context, 'date');
    });

    test('accepts a raw breakdown object (no envelope)', () {
      final b = TrustBreakdown.fromJson({
        'score': 50,
        'badge': 'Silver / Newbie',
        'rating_count': 0,
        'smoothed_rating': 2.5,
        'base_score': 50.0,
        'kyc_bonus': 0.0,
        'report_count': 0,
        'report_penalty': 0.0,
        'raw_score': 50.0,
      });
      expect(b.score, 50);
      expect(b.badge, 'Silver / Newbie');
      expect(b.recentRatings, isEmpty);
    });

    test('coerces int-shaped doubles defensively', () {
      // JSON deserializers sometimes hand us int for whole-number floats.
      final b = TrustBreakdown.fromJson({
        'breakdown': {
          'score': 80,
          'kyc_bonus': 10, // int, not 10.0
          'smoothed_rating': 4, // int
          'base_score': 80,
          'report_count': 0,
          'report_penalty': 0,
          'rating_count': 5,
          'raw_score': 90,
        },
      });
      expect(b.kycBonus, 10.0);
      expect(b.smoothedRating, 4.0);
      expect(b.baseScore, 80.0);
    });
  });

  group('TrustBreakdownScreen', () {
    testWidgets('renders score, badge, and component labels',
        (tester) async {
      const sample = TrustBreakdown(
        score: 72,
        badge: 'Gold / Member',
        ratingCount: 4,
        smoothedRating: 3.6,
        baseScore: 72.0,
        kycBonus: 10.0,
        reportCount: 1,
        reportPenalty: 15.0,
        rawScore: 67.0,
        recentRatings: [],
      );

      await tester.pumpWidget(
        ProviderScope(
          overrides: [
            trustBreakdownProvider
                .overrideWith((ref) => Future.value(sample)),
          ],
          child: const MaterialApp(home: TrustBreakdownScreen()),
        ),
      );
      // Let the AsyncValue resolve.
      await tester.pumpAndSettle(const Duration(seconds: 1));

      expect(find.text('Gold / Member'), findsOneWidget);
      // The chevron-style component rows must be visible.
      expect(find.textContaining('KYC'), findsOneWidget);
      expect(find.textContaining('Шағымдар'), findsOneWidget);
      expect(find.textContaining('Қорытынды'), findsOneWidget);
      // Final score value renders twice: in the big badge and in the
      // "Final" component row.
      expect(find.text('72'), findsNWidgets(2));
    });
  });
}
