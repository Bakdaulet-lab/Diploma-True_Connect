'use client';

import { useState } from 'react';
import { Heart, X, MapPin, Sparkles } from 'lucide-react';
import { useCandidates, useLikeUser, usePassUser } from '@/hooks/api';
import { LoadingScreen, EmptyState, TrustBadge, Avatar } from '@/components/ui/common';
import { calculateAge } from '@/lib/utils';
import type { MatchCandidate } from '@/types';
import toast from 'react-hot-toast';

export default function DiscoverPage() {
  const { data, fetchNextPage, hasNextPage, isLoading } = useCandidates();
  const likeMutation = useLikeUser();
  const passMutation = usePassUser();
  const [currentIndex, setCurrentIndex] = useState(0);
  const [showMatch, setShowMatch] = useState(false);

  // ИСПРАВЛЕНО: Безопасное извлечение массива из ответа API
  const candidates = data?.pages.flatMap((p: any) => p.data || p.items || p || []) ?? [];
  const current = candidates[currentIndex] as MatchCandidate | undefined;

  const handleLike = () => {
    if (!current) return;
    likeMutation.mutate(current.user_id, {
      onSuccess: (result) => {
        if (result.matched) setShowMatch(true);
        advance();
      },
      onError: () => toast.error('Failed to like'),
    });
  };

  const handlePass = () => {
    if (!current) return;
    passMutation.mutate(current.user_id, {
      onSuccess: () => advance(),
      onError: () => toast.error('Failed to pass'),
    });
  };

  const advance = () => {
    const nextIdx = currentIndex + 1;
    if (nextIdx >= candidates.length && hasNextPage) {
      fetchNextPage();
    }
    setCurrentIndex(nextIdx);
  };

  if (isLoading) return <LoadingScreen />;

  if (!current) {
    return (
      <EmptyState
        icon={<Sparkles className="h-12 w-12" />}
        title="No more people nearby"
        description="Check back later for new profiles"
      />
    );
  }

  return (
    <div className="flex flex-col items-center">
      <h1 className="mb-6 text-2xl font-bold text-gray-900">Discover</h1>

      {/* Profile Card */}
      <div className="relative w-full max-w-sm overflow-hidden rounded-3xl bg-white shadow-xl">
        {/* Photo */}
        <div className="relative h-[460px] bg-gray-200">
          {current.avatar_url ? (
            <img
              src={current.avatar_url}
              alt={current.display_name}
              className="h-full w-full object-cover"
            />
          ) : (
            <div className="flex h-full items-center justify-center">
              <Avatar name={current.display_name} size="xl" />
            </div>
          )}
          <div className="absolute inset-0 bg-gradient-to-t from-black/70 via-transparent to-transparent" />

          {/* Info overlay */}
          <div className="absolute bottom-0 left-0 right-0 p-6 text-white">
            <div className="flex items-center gap-2">
              <h2 className="text-2xl font-bold">{current.display_name}</h2>
              {current.birth_date && (
                <span className="text-xl">{calculateAge(current.birth_date)}</span>
              )}
            </div>
            {current.city && (
              <p className="mt-1 flex items-center gap-1 text-sm text-white/80">
                <MapPin className="h-4 w-4" />
                {current.city}
              </p>
            )}
            {current.bio && (
              <p className="mt-2 line-clamp-2 text-sm text-white/90">{current.bio}</p>
            )}
          </div>
        </div>

        {/* Actions */}
        <div className="flex items-center justify-center gap-6 p-6">
          <button
            onClick={handlePass}
            disabled={passMutation.isPending}
            className="flex h-16 w-16 items-center justify-center rounded-full border-2 border-gray-300 text-gray-400 transition-colors hover:border-red-400 hover:text-red-400"
          >
            <X className="h-8 w-8" />
          </button>
          <button
            onClick={handleLike}
            disabled={likeMutation.isPending}
            className="flex h-16 w-16 items-center justify-center rounded-full bg-primary-500 text-white shadow-lg transition-colors hover:bg-primary-600"
          >
            <Heart className="h-8 w-8" fill="currentColor" />
          </button>
        </div>
      </div>

      {/* Match Dialog */}
      {showMatch && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="w-full max-w-sm rounded-3xl bg-white p-8 text-center">
            <div className="mx-auto mb-4 flex h-20 w-20 items-center justify-center rounded-full bg-primary-100">
              <Heart className="h-10 w-10 text-primary-500" fill="currentColor" />
            </div>
            <h2 className="text-2xl font-bold text-gray-900">It&apos;s a Match!</h2>
            <p className="mt-2 text-sm text-gray-500">
              You and {current?.display_name} liked each other
            </p>
            <div className="mt-6 flex gap-3">
              <button onClick={() => setShowMatch(false)} className="btn-secondary flex-1">
                Keep Swiping
              </button>
              <a href="/matches" className="btn-primary flex-1">
                Send Message
              </a>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}