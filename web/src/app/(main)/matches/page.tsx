'use client';

import Link from 'next/link';
import { MessageCircle } from 'lucide-react';
import { useMatches } from '@/hooks/api';
import { LoadingScreen, EmptyState, Avatar } from '@/components/ui/common';
import { timeAgo } from '@/lib/utils';

export default function MatchesPage() {
  const { data: matches, isLoading } = useMatches();

  if (isLoading) return <LoadingScreen />;

  return (
    <div>
      <h1 className="mb-6 text-2xl font-bold text-gray-900">Matches</h1>

      {!matches?.length ? (
        <EmptyState
          icon={<MessageCircle className="h-12 w-12" />}
          title="No matches yet"
          description="Start swiping to find your match!"
          action={
            <Link href="/discover" className="btn-primary">
              Discover People
            </Link>
          }
        />
      ) : (
        <div className="space-y-2">
          {matches.map((match) => (
            <Link
              key={match.id}
              href={`/chat/${match.id}`}
              className="flex items-center gap-4 rounded-2xl bg-white p-4 transition-colors hover:bg-gray-50"
            >
              <Avatar
                src={match.other_user?.avatar_url}
                name={match.other_user?.display_name || 'User'}
                size="lg"
              />
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between">
                  <h3 className="font-semibold text-gray-900 truncate">
                    {match.other_user?.display_name || 'Unknown'}
                  </h3>
                  {match.last_message && (
                    <span className="text-xs text-gray-400">
                      {timeAgo(match.last_message.created_at)}
                    </span>
                  )}
                </div>
                <p className="mt-0.5 truncate text-sm text-gray-500">
                  {match.last_message?.content || 'Say hello! 👋'}
                </p>
              </div>
            </Link>
          ))}
        </div>
      )}
    </div>
  );
}
