import 'package:flutter/services.dart';

/// Formats phone input as +7 XXX XXX XX XX (Kazakhstan format).
/// Strips spaces when reading the raw value for the backend: controller.text.replaceAll(' ', '').
class KzPhoneFormatter extends TextInputFormatter {
  @override
  TextEditingValue formatEditUpdate(
    TextEditingValue oldValue,
    TextEditingValue newValue,
  ) {
    if (!newValue.text.startsWith('+7')) {
      return const TextEditingValue(
        text: '+7',
        selection: TextSelection.collapsed(offset: 2),
      );
    }

    // Extract the digit sequence after '+': e.g. "+7 701 234 56 78" → "77012345678"
    final afterPlus = newValue.text.substring(1);
    final allDigits = afterPlus.replaceAll(RegExp(r'\D'), '');

    // The first digit is always '7' from the '+7' prefix; the remaining are user digits.
    final userDigits = allDigits.length > 1 ? allDigits.substring(1) : '';
    final capped =
        userDigits.length > 10 ? userDigits.substring(0, 10) : userDigits;

    // Build "+7 XXX XXX XX XX"
    final buf = StringBuffer('+7');
    for (var i = 0; i < capped.length; i++) {
      if (i == 0 || i == 3 || i == 6 || i == 8) buf.write(' ');
      buf.write(capped[i]);
    }

    final formatted = buf.toString();
    return TextEditingValue(
      text: formatted,
      selection: TextSelection.collapsed(offset: formatted.length),
    );
  }
}
