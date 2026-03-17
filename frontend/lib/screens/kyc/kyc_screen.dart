import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:image_picker/image_picker.dart';
import '../../core/theme/app_colors.dart';
import '../../core/constants/app_strings.dart';
import '../../providers/kyc_provider.dart';
import '../../services/api_service.dart';

class KycScreen extends ConsumerWidget {
  const KycScreen({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final kycAsync = ref.watch(kycStatusProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text(AppStrings.verification),
      ),
      body: kycAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (e, _) => Center(child: Text('Error: $e')),
        data: (status) => SingleChildScrollView(
          padding: const EdgeInsets.all(24),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // Status card
              Card(
                child: Padding(
                  padding: const EdgeInsets.all(24),
                  child: Column(
                    children: [
                      Icon(
                        _statusIcon(status.status),
                        size: 64,
                        color: _statusColor(status.status),
                      ),
                      const SizedBox(height: 16),
                      Text(
                        _statusTitle(status.status),
                        style: const TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        _statusDescription(status.status),
                        textAlign: TextAlign.center,
                        style: TextStyle(color: Colors.grey[600]),
                      ),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 32),
              if (status.isNone || status.isRejected) ...[
                const Text(
                  'Submit a government-issued ID document:',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.w500,
                  ),
                ),
                const SizedBox(height: 8),
                Text(
                  '• Passport, national ID, or driver\'s license\n'
                  '• JPEG, PNG, or PDF format\n'
                  '• Maximum 10MB file size',
                  style: TextStyle(color: Colors.grey[600], height: 1.5),
                ),
                const SizedBox(height: 24),
                ElevatedButton.icon(
                  onPressed: () => _uploadDocument(context, ref),
                  icon: const Icon(Icons.upload_file),
                  label: const Text(AppStrings.submitDocument),
                ),
                const SizedBox(height: 12),
                OutlinedButton.icon(
                  onPressed: () => _takePhoto(context, ref),
                  icon: const Icon(Icons.camera_alt),
                  label: const Text('Take Photo'),
                ),
              ],
              if (status.isPending)
                const Card(
                  child: Padding(
                    padding: EdgeInsets.all(16),
                    child: Row(
                      children: [
                        CircularProgressIndicator(strokeWidth: 2),
                        SizedBox(width: 16),
                        Expanded(
                          child: Text(
                            'Your document is being reviewed. This usually takes 1-3 business days.',
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
    );
  }

  Future<void> _uploadDocument(BuildContext context, WidgetRef ref) async {
    final picker = ImagePicker();
    final image = await picker.pickImage(
      source: ImageSource.gallery,
      maxWidth: 2048,
      maxHeight: 2048,
      imageQuality: 90,
    );
    if (image != null) {
      try {
        await ref.read(apiServiceProvider).submitKyc(image.path);
        ref.invalidate(kycStatusProvider);
        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Document submitted successfully')),
          );
        }
      } catch (e) {
        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Error: $e')),
          );
        }
      }
    }
  }

  Future<void> _takePhoto(BuildContext context, WidgetRef ref) async {
    final picker = ImagePicker();
    final image = await picker.pickImage(
      source: ImageSource.camera,
      maxWidth: 2048,
      maxHeight: 2048,
      imageQuality: 90,
    );
    if (image != null) {
      try {
        await ref.read(apiServiceProvider).submitKyc(image.path);
        ref.invalidate(kycStatusProvider);
        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Document submitted successfully')),
          );
        }
      } catch (e) {
        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text('Error: $e')),
          );
        }
      }
    }
  }

  IconData _statusIcon(String status) {
    switch (status) {
      case 'approved':
        return Icons.verified;
      case 'pending':
        return Icons.hourglass_top;
      case 'rejected':
        return Icons.cancel;
      default:
        return Icons.badge_outlined;
    }
  }

  Color _statusColor(String status) {
    switch (status) {
      case 'approved':
        return AppColors.success;
      case 'pending':
        return AppColors.warning;
      case 'rejected':
        return AppColors.error;
      default:
        return AppColors.textSecondaryLight;
    }
  }

  String _statusTitle(String status) {
    switch (status) {
      case 'approved':
        return AppStrings.verificationApproved;
      case 'pending':
        return AppStrings.verificationPending;
      case 'rejected':
        return 'Verification Rejected';
      default:
        return 'Not Verified';
    }
  }

  String _statusDescription(String status) {
    switch (status) {
      case 'approved':
        return 'Your identity has been verified. You now have full access to all features.';
      case 'pending':
        return 'Your document is being reviewed by our team.';
      case 'rejected':
        return 'Your document was rejected. Please submit a clearer photo.';
      default:
        return 'Verify your identity to build trust and unlock all features.';
    }
  }
}
