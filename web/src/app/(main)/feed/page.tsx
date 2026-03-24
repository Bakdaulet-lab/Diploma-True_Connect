'use client';

import { useState, useRef, useCallback } from 'react';
import { Plus, Newspaper } from 'lucide-react';
import { usePosts } from '@/hooks/api';
import { LoadingScreen, EmptyState } from '@/components/ui/common';
import { PostCard } from '@/components/feed/post-card';
import { CreatePostModal } from '@/components/feed/create-post-modal';

export default function FeedPage() {
  const { data, fetchNextPage, hasNextPage, isFetchingNextPage, isLoading } = usePosts();
  const [showCreatePost, setShowCreatePost] = useState(false);

  const posts = data?.pages.flatMap((p) => p.items) ?? [];

  // Infinite scroll observer
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
      <div className="mb-6 flex items-center justify-between">
        <h1 className="text-2xl font-bold text-gray-900">Feed</h1>
        <button onClick={() => setShowCreatePost(true)} className="btn-primary gap-2">
          <Plus className="h-4 w-4" />
          New Post
        </button>
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
