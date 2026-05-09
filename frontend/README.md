# TrueConnect Flutter Frontend

Mobile frontend for the TrueConnect trust-based social networking and dating platform.

## Tech Stack

- **Flutter 3.16+** / Dart 3.2+
- **State Management:** Riverpod
- **Navigation:** GoRouter
- **HTTP Client:** Dio (with JWT interceptor)
- **WebSocket:** web_socket_channel (real-time chat)
- **Secure Storage:** flutter_secure_storage

## Project Structure

```
lib/
├── main.dart                    # App entry point
├── core/
│   ├── constants/               # API endpoints, strings
│   ├── network/                 # Dio client, JWT interceptor, exceptions
│   ├── router/                  # GoRouter configuration
│   ├── storage/                 # Secure token storage
│   └── theme/                   # Material 3 theme, colors
├── models/                      # Data models (User, Profile, Post, Match, etc.)
├── providers/                   # Riverpod state management
│   ├── auth_provider.dart       # Authentication state
│   ├── profile_provider.dart    # Profile management
│   ├── matching_provider.dart   # Discovery & matching
│   ├── chat_provider.dart       # Real-time messaging
│   ├── feed_provider.dart       # Social feed
│   ├── settings_provider.dart   # User settings
│   └── kyc_provider.dart        # Identity verification
├── services/
│   ├── api_service.dart         # HTTP API client (all endpoints)
│   └── websocket_service.dart   # WebSocket for real-time chat
├── screens/
│   ├── splash/                  # Splash screen
│   ├── auth/                    # Login & Register
│   ├── home/                    # Bottom navigation shell
│   ├── discovery/               # Swipe matching
│   ├── matches/                 # Match list
│   ├── chat/                    # Chat messaging
│   ├── feed/                    # Social feed & create post
│   ├── profile/                 # Profile view, edit, other users
│   ├── settings/                # App settings
│   └── kyc/                     # Identity verification
└── widgets/                     # Reusable components
    ├── profile_card.dart        # Swipeable profile card
    ├── post_card.dart           # Feed post card
    ├── message_bubble.dart      # Chat message bubble
    ├── photo_gallery.dart       # Photo grid with delete
    ├── empty_state.dart         # Empty state placeholder
    └── common_widgets.dart      # Loading overlay, trust badge
```

## Features

- **Authentication** — Phone + password login/register with JWT token management
- **Discovery** — Swipe-based matching with like/pass actions
- **Matches** — List of mutual matches with last message preview
- **Real-time Chat** — WebSocket-powered messaging with typing indicators
- **Social Feed** — Create posts, like, comment with infinite scroll
- **Profile Management** — Edit profile, upload/delete photos
- **Trust System** — Trust score badges with color-coded levels
- **Identity Verification (KYC)** — Document upload via camera or gallery
- **Settings** — Distance, age range, notifications, privacy controls
- **Reporting** — Report users with reason input

## Getting Started

### Prerequisites

- Flutter SDK 3.16+
- Dart SDK 3.2+
- Android Studio / Xcode (for mobile emulators)

### Setup

```bash
cd frontend

# Install dependencies
flutter pub get

# Run on connected device or emulator
flutter run

# Run on specific platform
flutter run -d chrome     # Web
flutter run -d android    # Android
flutter run -d ios        # iOS
```

### Backend Connection

The app connects to the backend at `http://localhost:8080` by default. Update `lib/core/constants/api_constants.dart` to change the API URL.

## Architecture

- **Clean Architecture** — Models → Services → Providers → Screens
- **Riverpod** — StateNotifier for complex state, FutureProvider for simple async data
- **GoRouter** — Declarative routing with auth redirect guards
- **Dio Interceptors** — Automatic JWT token injection and refresh on 401
- **WebSocket** — Auto-reconnect with ping/pong keepalive
