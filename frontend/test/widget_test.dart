import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:trueconnect/main.dart';

void main() {
  testWidgets('App smoke test', (WidgetTester tester) async {
    // Build our app and trigger a frame.
    await tester.pumpWidget(const MyApp());

    // Verify that our app displays 'Coming Soon'
    expect(find.text('TrueConnect - Coming Soon'), findsOneWidget);
  });
}
