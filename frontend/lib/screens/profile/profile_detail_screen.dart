import 'package:flutter/material.dart';
import 'package:haptic_feedback/haptic_feedback.dart';

class ProfileDetailScreen extends StatelessWidget {
  final Map<String, dynamic> profile;

  const ProfileDetailScreen({super.key, required this.profile});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      extendBodyBehindAppBar: true,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back_ios, color: Colors.white, shadows: [Shadow(blurRadius: 4, color: Colors.black54)]),
          onPressed: () {
            Haptics.vibrate(HapticsType.selection);
            Navigator.pop(context);
          },
        ),
      ),
      body: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Hero(
              tag: 'profile-img-${profile["id"]}',
              child: Image.network(
                profile['imageUrl'],
                height: 500,
                fit: BoxFit.cover,
                loadingBuilder: (context, child, progress) {
                  if (profile['imageUrl'] == 'https://via.placeholder.com/500') {
                    return Container(
                      height: 500,
                      color: Colors.grey[300],
                      alignment: Alignment.center,
                      child: const Icon(Icons.person, size: 100, color: Colors.white),
                    );
                  }
                  if (progress == null) return child;
                  return SizedBox(
                    height: 500,
                    child: Center(
                      child: CircularProgressIndicator(
                        value: progress.expectedTotalBytes != null
                            ? progress.cumulativeBytesLoaded / progress.expectedTotalBytes!
                            : null,
                      ),
                    ),
                  );
                },
                errorBuilder: (context, error, stackTrace) => Container(
                  height: 500,
                  color: Colors.grey[300],
                  alignment: Alignment.center,
                  child: const Icon(Icons.person, size: 100, color: Colors.white),
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(24.0),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Expanded(
                        child: Text(
                          '${profile["name"]}, ${profile["age"]}',
                          style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                                fontWeight: FontWeight.bold,
                              ),
                        ),
                      ),
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                        decoration: BoxDecoration(
                          color: Colors.deepPurple.shade50,
                          borderRadius: BorderRadius.circular(20),
                        ),
                        child: Row(
                          children: [
                            const Icon(Icons.verified_user, color: Colors.deepPurple, size: 16),
                            const SizedBox(width: 4),
                            Text(
                              'Score: ${profile["trustScore"]}',
                              style: const TextStyle(
                                color: Colors.deepPurple,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      const Icon(Icons.location_on, color: Colors.grey, size: 18),
                      const SizedBox(width: 4),
                      Text(
                        profile['city'],
                        style: const TextStyle(color: Colors.grey, fontSize: 16),
                      ),
                    ],
                  ),
                  const SizedBox(height: 24),
                  const Text(
                    'About',
                    style: TextStyle(
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 12),
                  Text(
                    profile['bio'] ?? 'No bio provided.',
                    style: const TextStyle(
                      fontSize: 16,
                      color: Colors.black87,
                      height: 1.5,
                    ),
                  ),
                  if (profile['prompts'] != null && (profile['prompts'] as List).isNotEmpty) ...[
                    const SizedBox(height: 24),
                    const Text(
                      'Prompts',
                      style: TextStyle(
                        fontSize: 20,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 12),
                    ...((profile['prompts'] as List).map((prompt) {
                      return Container(
                        width: double.infinity,
                        margin: const EdgeInsets.only(bottom: 12),
                        padding: const EdgeInsets.all(16),
                        decoration: BoxDecoration(
                          color: Theme.of(context).primaryColor.withValues(alpha: 0.05),
                          borderRadius: BorderRadius.circular(12),
                          border: Border.all(
                            color: Theme.of(context).primaryColor.withValues(alpha: 0.2),
                          ),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              prompt['question'] ?? '',
                              style: const TextStyle(
                                fontSize: 14,
                                color: Colors.grey,
                                fontStyle: FontStyle.italic,
                              ),
                            ),
                            const SizedBox(height: 8),
                            Text(
                              prompt['answer'] ?? '',
                              style: const TextStyle(
                                fontSize: 16,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                          ],
                        ),
                      );
                    }).toList()),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
      floatingActionButton: Padding(
        padding: const EdgeInsets.only(bottom: 20.0),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceEvenly,
          children: [
            const Spacer(),
            FloatingActionButton(
              heroTag: 'pass-btn',
              onPressed: () {
                Haptics.vibrate(HapticsType.heavy);
                Navigator.pop(context, 'pass');
              },
              backgroundColor: Colors.white,
              elevation: 8,
              child: const Icon(Icons.close, color: Colors.red, size: 32),
            ),
            const SizedBox(width: 40),
            FloatingActionButton(
              heroTag: 'match-btn',
              onPressed: () {
                Haptics.vibrate(HapticsType.success);
                Navigator.pop(context, 'like');
              },
              backgroundColor: Colors.deepPurple,
              elevation: 8,
              child: const Icon(Icons.favorite, color: Colors.white, size: 32),
            ),
            const Spacer(),
          ],
        ),
      ),
      floatingActionButtonLocation: FloatingActionButtonLocation.centerDocked,
    );
  }
}
