class Venue {
  final String id;
  final String name;
  final String city;
  final String category; // cafe | restaurant | park | cultural
  final String address;
  final String? phone;

  const Venue({
    required this.id,
    required this.name,
    required this.city,
    required this.category,
    required this.address,
    this.phone,
  });

  factory Venue.fromJson(Map<String, dynamic> json) => Venue(
        id: json['id'] as String? ?? '',
        name: json['name'] as String? ?? '',
        city: json['city'] as String? ?? '',
        category: json['category'] as String? ?? 'cafe',
        address: json['address'] as String? ?? '',
        phone: json['phone'] as String?,
      );

  String get categoryLabel {
    switch (category) {
      case 'restaurant':
        return 'Мейрамхана';
      case 'park':
        return 'Саябақ';
      case 'cultural':
        return 'Мәдени орын';
      default:
        return 'Кафе';
    }
  }

  String get categoryEmoji {
    switch (category) {
      case 'restaurant':
        return '🍽️';
      case 'park':
        return '🌿';
      case 'cultural':
        return '🏛️';
      default:
        return '☕';
    }
  }
}
