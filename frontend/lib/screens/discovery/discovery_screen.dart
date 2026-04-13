import 'package:flutter/material.dart';
import 'package:flutter_card_swiper/flutter_card_swiper.dart';
import 'package:haptic_feedback/haptic_feedback.dart';
import 'package:trueconnect/screens/profile/profile_detail_screen.dart';

class DiscoveryScreen extends StatefulWidget {
  const DiscoveryScreen({super.key});

  @override
  State<DiscoveryScreen> createState() => _DiscoveryScreenState();
}

class _DiscoveryScreenState extends State<DiscoveryScreen> {
  final CardSwiperController _controller = CardSwiperController();

  final List<Map<String, dynamic>> _candidates = [
    {
      'id': '1',
      'name': 'Alicia',
      'age': 25,
      'city': 'New York',
      'trustScore': 95,
      'imageUrl': 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?ixlib=rb-1.2.1&auto=format&fit=crop&w=800&q=80',
      'bio': 'Lover of art, coffee, and all things dogs. Looking for a genuine connection.',
      'prompts': [
        {
          'question': 'My simple pleasure in life...',
          'answer': 'Morning coffee walks in Central Park.',
        },
        {
          'question': 'Two truths and a lie...',
          'answer': 'I love dogs, I hate pizza, I once met a celebrity.',
        },
      ],
    },
    {
      'id': '2',
      'name': 'David',
      'age': 31,
      'city': 'Los Angeles',
      'trustScore': 88,
      'imageUrl': 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?ixlib=rb-1.2.1&auto=format&fit=crop&w=800&q=80',
      'bio': "Software engineer by day, passionate chef by night. Let's grab pasta.",
      'prompts': [
        {
          'question': 'I go crazy for...',
          'answer': 'A perfectly cooked carbonara.',
        },
      ],
    },
    {
      'id': '3',
      'name': 'Emma',
      'age': 28,
      'city': 'Chicago',
      'trustScore': 75,
      'imageUrl': 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?ixlib=rb-1.2.1&auto=format&fit=crop&w=800&q=80',
      'bio': 'Adventure seeker. Frequent flyer. Looking to explore the world.'
    },
  ];

  bool _onSwipe(
    int previousIndex,
    int? currentIndex,
    CardSwiperDirection direction,
  ) {
    if (direction == CardSwiperDirection.right) {
      // Like
      Haptics.vibrate(HapticsType.success);
    } else if (direction == CardSwiperDirection.left) {
      // Pass
      Haptics.vibrate(HapticsType.heavy);
    } else if (direction == CardSwiperDirection.top) {
      // Super Like
      Haptics.vibrate(HapticsType.light);
    }
    return true;
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  void _onCardTap(int index) async {
    Haptics.vibrate(HapticsType.selection);
    
    // Complex Hero transition
    final result = await Navigator.push(
      context,
      PageRouteBuilder(
        transitionDuration: const Duration(milliseconds: 500), // Smooth micro-interaction
        pageBuilder: (context, animation, secondaryAnimation) {
          return FadeTransition(
            opacity: animation,
            child: ProfileDetailScreen(profile: _candidates[index]),
          );
        },
      ),
    );

    if (result == 'like') {
      _controller.swipeRight();
    } else if (result == 'pass') {
      _controller.swipeLeft();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey[100],
      appBar: AppBar(
        title: const Text('TrueConnect', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 24, color: Colors.deepPurple)),
        elevation: 0,
        backgroundColor: Colors.transparent,
        centerTitle: true,
      ),
      body: SafeArea(
        child: Column(
          children: [
            Expanded(
              child: _candidates.isEmpty
                  ? const Center(child: Text('No more candidates nearby'))
                  : CardSwiper(
                      controller: _controller,
                      cardsCount: _candidates.length,
                      onSwipe: _onSwipe,
                      isLoop: false,
                      numberOfCardsDisplayed: _candidates.length > 2 ? 3 : _candidates.length,
                      backCardOffset: const Offset(0, 40),
                      padding: const EdgeInsets.all(24.0),
                      cardBuilder: (
                        context,
                        index,
                        horizontalOffsetPercentage,
                        verticalOffsetPercentage,
                      ) {
                        final candidate = _candidates[index];
                        return _buildCard(candidate, index);
                      },
                    ),
            ),
            Padding(
              padding: const EdgeInsets.only(bottom: 30.0, top: 10),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  FloatingActionButton(
                    heroTag: 'list-pass',
                    onPressed: () => _controller.swipeLeft(),
                    backgroundColor: Colors.white,
                    child: const Icon(Icons.close, color: Colors.red, size: 30),
                  ),
                  FloatingActionButton(
                    heroTag: 'list-like',
                    onPressed: () => _controller.swipeRight(),
                    backgroundColor: Colors.white,
                    child: const Icon(Icons.favorite, color: Colors.green, size: 30),
                  ),
                  FloatingActionButton(
                    heroTag: 'list-star',
                    onPressed: () => _controller.swipeTop(),
                    backgroundColor: Colors.white,
                    child: const Icon(Icons.star, color: Colors.blue, size: 30),
                  ),
                ],
              ),
            )
          ],
        ),
      ),
    );
  }

  Widget _buildCard(Map<String, dynamic> candidate, int index) {
    return GestureDetector(
      onTap: () => _onCardTap(index),
      child: Hero(
        tag: 'profile-img-${candidate["id"]}', // Matching hero tag
        child: ClipRRect(
          borderRadius: BorderRadius.circular(20),
          child: Stack(
            fit: StackFit.expand,
            children: [
              Image.network(
                candidate['imageUrl'],
                fit: BoxFit.cover,
              ),
              Container(
                decoration: BoxDecoration(
                  gradient: LinearGradient(
                    colors: [
                      Colors.transparent,
                      Colors.black.withValues(alpha: 0.8),
                    ],
                    begin: Alignment.topCenter,
                    end: Alignment.bottomCenter,
                    stops: const [0.6, 1.0],
                  ),
                ),
              ),
              Positioned(
                bottom: 20,
                left: 20,
                right: 20,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Text(
                          '${candidate["name"]}, ${candidate["age"]}',
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 26,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(width: 8),
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                          decoration: BoxDecoration(
                            color: Colors.white24,
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Text(
                            'Score: ${candidate["trustScore"]}',
                            style: const TextStyle(color: Colors.white, fontWeight: FontWeight.bold),
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        const Icon(Icons.location_on, color: Colors.white70, size: 16),
                        const SizedBox(width: 4),
                        Text(
                          candidate['city'],
                          style: const TextStyle(
                            color: Colors.white70,
                            fontSize: 16,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
