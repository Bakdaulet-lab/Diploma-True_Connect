'use client';

import { useState, useRef, useCallback } from 'react';
import { Plus, Newspaper, Search, ListFilter } from 'lucide-react';
import { usePosts } from '@/hooks/api';
import { LoadingScreen, EmptyState } from '@/components/ui/common';
import { PostCard } from '@/components/feed/post-card';
import { CreatePostModal } from '@/components/feed/create-post-modal';

export default function FeedPage() {
  const [showCreatePost, setShowCreatePost] = useState(false);
  
  const [filterState, setFilterState] = useState({ query: '', sort: 'recent', timeframe: 'all' });
  const [localQuery, setLocalQuery] = useState('');

  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isLoading } = usePosts(filterState);

// Теперь фронтенд найдет посты, где бы они ни прятались в ответе
const posts = data?.pages?.flatMap((p: any) => Array.isArray(p) ? p : (p.items || p.data || [])) ?? [];  // Infinite scroll observer
  const observer = useRef<IntersectionObserver>();
  const lastPostRef = useCallback(
    (node: HTMLDivElement | null) => {
      if (isFetchingNextPage) return;
      if (observer.current) observer.current.disconnect();
      observer.current = new IntersectionObserver((entries) => {
        if (entries[0].isIntersecting && hasNextPage) {
          fetchNextPage();
        }
      });
      if (node) observer.current.observe(node);
    },
    [isFetchingNextPage, hasNextPage, fetchNextPage]
  );

  if (isLoading) return <LoadingScreen />;

  return (
    <div>
      <div className="mb-6 flex flex-col gap-4">
        <div className="flex items-center justify-between">
          <h1 className="text-2xl font-bold text-gray-900">Feed</h1>
          <button onClick={() => setShowCreatePost(true)} className="btn-primary gap-2">
            <Plus className="h-4 w-4" />
            New Post
          </button>
        </div>

        <div className="flex flex-col sm:flex-row gap-3 bg-white p-4 rounded-xl shadow-sm border border-gray-100">
          <div className="relative flex-1">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
            <input
              type="text"
              placeholder="Search posts..."
              className="w-full pl-10 pr-4 py-2 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-primary-500"
              value={localQuery}
              onChange={(e) => setLocalQuery(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') setFilterState(prev => ({ ...prev, query: localQuery }));
              }}
            />
          </div>
          <div className="flex items-center gap-2">
            <ListFilter className="h-4 w-4 text-gray-400" />
            <select
              className="py-2 px-3 border border-gray-200 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              value={filterState.sort}
              onChange={(e) => setFilterState(prev => ({ ...prev, sort: e.target.value }))}
            >
              <option value="recent">Recent</option>
              <option value="popular">Popular</option>
            </select>
            <select
              className="py-2 px-3 border border-gray-200 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-primary-500"
              value={filterState.timeframe}
              onChange={(e) => setFilterState(prev => ({ ...prev, timeframe: e.target.value }))}
            >
              <option value="all">All Time</option>
              <option value="24h">Past 24h</option>
              <option value="7d">Past 7 days</option>
              <option value="30d">Past 30 days</option>
            </select>
          </div>
        </div>
      </div>

      {!posts.length ? (
        <EmptyState
          icon={<Newspaper className="h-12 w-12" />}
          title="No posts yet"
          description="Be the first to share something!"
          action={
            <button onClick={() => setShowCreatePost(true)} className="btn-primary">
              Create Post
            </button>
          }
        />
      ) : (
        <div className="space-y-4">
          {posts.map((post, i) => (
            <div key={post.id} ref={i === posts.length - 1 ? lastPostRef : undefined}>
              <PostCard post={post} />
            </div>
          ))}
          {isFetchingNextPage && (
            <div className="py-4 text-center text-sm text-gray-400">Loading more...</div>
          )}
        </div>
      )}

      {showCreatePost && <CreatePostModal onClose={() => setShowCreatePost(false)} />}
    </div>
  );
}
