import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/models.dart';
import '../services/api_service.dart';

class MatchingState {
  final List<Profile> candidates;
  final List<Match> matches;
  final bool isLoading;
  final bool hasMoreCandidates;
  final String? error;
  final Match? newMatch;

  const MatchingState({
    this.candidates = const [],
    this.matches = const [],
    this.isLoading = false,
    this.hasMoreCandidates = true,
    this.error,
    this.newMatch,
  });

  MatchingState copyWith({
    List<Profile>? candidates,
    List<Match>? matches,
    bool? isLoading,
    bool? hasMoreCandidates,
    String? error,
    Match? newMatch,
  }) =>
      MatchingState(
        candidates: candidates ?? this.candidates,
        matches: matches ?? this.matches,
        isLoading: isLoading ?? this.isLoading,
        hasMoreCandidates: hasMoreCandidates ?? this.hasMoreCandidates,
        error: error,
        newMatch: newMatch,
      );
}

class MatchingNotifier extends StateNotifier<MatchingState> {
  final ApiService _apiService;
  int _page = 1;

  MatchingNotifier(this._apiService) : super(const MatchingState());

  Future<void> loadCandidates({bool refresh = false}) async {
    if (refresh) _page = 1;
    if (state.isLoading) return;

    state = state.copyWith(isLoading: true, error: null);
    try {
      final candidates = await _apiService.getCandidates(page: _page);
      state = state.copyWith(
        candidates: refresh ? candidates : [...state.candidates, ...candidates],
        isLoading: false,
        hasMoreCandidates: candidates.length >= 10,
      );
      _page++;
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> likeUser(String targetId) async {
    try {
      final result = await _apiService.likeUser(targetId);
      state = state.copyWith(
        candidates:
            state.candidates.where((c) => c.userId != targetId).toList(),
      );
      // Check if it's a mutual match
      if (result['data'] != null && result['data']['matched'] == true) {
        final match = Match.fromJson(result['data']);
        state = state.copyWith(
          newMatch: match,
          matches: [match, ...state.matches],
        );
      }
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }

  Future<void> passUser(String targetId) async {
    try {
      await _apiService.passUser(targetId);
      state = state.copyWith(
        candidates:
            state.candidates.where((c) => c.userId != targetId).toList(),
      );
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }

  Future<void> loadMatches({bool refresh = false}) async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final matches = await _apiService.getMatches();
      state = state.copyWith(matches: matches, isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  void clearNewMatch() {
    state = state.copyWith(newMatch: null);
  }
}

final matchingProvider =
    StateNotifierProvider<MatchingNotifier, MatchingState>((ref) {
  return MatchingNotifier(ref.read(apiServiceProvider));
});
