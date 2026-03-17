class ApiException implements Exception {
  final String code;
  final String message;
  final Map<String, dynamic>? details;
  final int? statusCode;

  const ApiException({
    required this.code,
    required this.message,
    this.details,
    this.statusCode,
  });

  factory ApiException.fromJson(Map<String, dynamic> json, {int? statusCode}) {
    final error = json['error'] ?? json;
    return ApiException(
      code: error['code'] ?? 'UNKNOWN',
      message: error['message'] ?? 'An unknown error occurred',
      details: error['details'] != null
          ? Map<String, dynamic>.from(error['details'])
          : null,
      statusCode: statusCode,
    );
  }

  factory ApiException.network() => const ApiException(
        code: 'NETWORK_ERROR',
        message: 'No internet connection. Please check your network.',
      );

  factory ApiException.timeout() => const ApiException(
        code: 'TIMEOUT',
        message: 'Request timed out. Please try again.',
      );

  factory ApiException.unknown([String? message]) => ApiException(
        code: 'UNKNOWN',
        message: message ?? 'An unexpected error occurred.',
      );

  bool get isUnauthorized => statusCode == 401;
  bool get isForbidden => statusCode == 403;
  bool get isNotFound => statusCode == 404;
  bool get isValidation => code == 'VALIDATION_FAILED';

  @override
  String toString() => 'ApiException($code): $message';
}
