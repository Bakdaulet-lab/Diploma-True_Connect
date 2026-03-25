'use client';

import { useState } from 'react';
import { Heart, MessageCircle, Trash2, MoreHorizontal } from 'lucide-react';
import { useAuthStore } from '@/store/auth-store';
import { useLikePost, useDeletePost, useComments, useAddComment } from '@/hooks/api';
import { Avatar } from '@/components/ui/common';
import { timeAgo, cn } from '@/lib/utils';
import type { Post } from '@/types';
import toast from 'react-hot-toast';

// Формируем полные URL для MinIO
const getFullUrl = (path?: string) => {
  if (!path) return undefined;
  if (path.startsWith('http')) return path;
  const cleanPath = path.replace(/^\/+/, '');
  return `http://localhost:9000/trueconnect/${cleanPath}`;
};

export function PostCard({ post }: { post: Post }) {
  const userId = useAuthStore((s) => s.userId);
  const likeMutation = useLikePost();
  const deleteMutation = useDeletePost();
  const [showComments, setShowComments] = useState(false);
  const [showMenu, setShowMenu] = useState(false);
  const isOwner = post.author_id === userId;

  const handleLike = () => {
    likeMutation.mutate({ postId: post.id, liked: post.is_liked });
  };

  const handleDelete = () => {
    if (confirm('Delete this post?')) {
      deleteMutation.mutate(post.id, {
        onError: () => toast.error('Failed to delete post'),
      });
    }
  };

  const displayName = post.author_name || post.author?.display_name || 'Anonymous';
  const avatarUrl = getFullUrl(post.author_avatar || post.author?.avatar_url);

  return (
    <div className="card">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <Avatar src={avatarUrl} name={displayName} size="md" />
          <div>
            <p className="text-sm font-semibold text-gray-900">{displayName}</p>
            <p className="text-xs text-gray-400">{timeAgo(post.created_at)}</p>
          </div>
        </div>
        {isOwner && (
          <div className="relative">
            <button
              onClick={() => setShowMenu(!showMenu)}
              className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100 hover:text-gray-600"
            >
              <MoreHorizontal className="h-5 w-5" />
            </button>
            {showMenu && (
              <div className="absolute right-0 top-full z-10 mt-1 w-36 rounded-xl bg-white py-1 shadow-lg ring-1 ring-gray-200">
                <button
                  onClick={handleDelete}
                  className="flex w-full items-center gap-2 px-4 py-2 text-sm text-red-500 hover:bg-red-50"
                >
                  <Trash2 className="h-4 w-4" />
                  Delete
                </button>
              </div>
            )}
          </div>
        )}
      </div>

      {/* Content */}
      <p className="mt-3 text-sm text-gray-800 whitespace-pre-wrap">{post.content}</p>

      {/* Media */}
      {post.media_url && post.media_url !== "" && (
        <img
          src={getFullUrl(post.media_url)}
          alt="Post attachment"
          className="mt-3 w-full rounded-xl object-cover max-h-96"
        />
      )}

      {/* Actions */}
      <div className="mt-4 flex items-center gap-4 border-t border-gray-100 pt-3">
        <button
          onClick={handleLike}
          className={cn(
            'flex items-center gap-1.5 text-sm transition-colors',
            post.is_liked ? 'text-primary-500' : 'text-gray-500 hover:text-primary-500'
          )}
        >
          <Heart className="h-5 w-5" fill={post.is_liked ? 'currentColor' : 'none'} />
          {post.like_count}
        </button>
        <button
          onClick={() => setShowComments(!showComments)}
          className="flex items-center gap-1.5 text-sm text-gray-500 hover:text-primary-500 transition-colors"
        >
          <MessageCircle className="h-5 w-5" />
          {post.comment_count}
        </button>
      </div>

      {/* Comments Section */}
      {showComments && <CommentsSection postId={post.id} />}
    </div>
  );
}

function CommentsSection({ postId }: { postId: string }) {
  const { data, isLoading } = useComments(postId);
  const addComment = useAddComment();
  const [text, setText] = useState('');

  // КЛЮЧЕВОЕ ИСПРАВЛЕНИЕ: Берем данные из поля "items", так как бэкенд отдает именно его
  const rawData: any = data;
  const commentsList = rawData?.items || [];

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!text.trim()) return;
    addComment.mutate(
      { postId, content: text.trim() },
      {
        onSuccess: () => {
          setText('');
          toast.success('Comment added');
        },
        onError: () => toast.error('Failed to add comment'),
      }
    );
  };

  return (
    <div className="mt-3 border-t border-gray-100 pt-3">
      {isLoading ? (
        <p className="text-sm text-gray-400">Loading comments...</p>
      ) : (
        <div className="space-y-3 max-h-60 overflow-y-auto pr-1">
          {commentsList.length > 0 ? (
            commentsList.map((comment: any) => {
              const commentAuthor = comment.author_name || 'User';
              const commentAvatar = getFullUrl(comment.author_avatar || comment.author?.avatar_url);
              
              return (
                <div key={comment.id} className="flex gap-2 text-left">
                  <Avatar src={commentAvatar} name={commentAuthor} size="sm" />
                  <div className="flex-1">
                    <p className="text-xs font-semibold text-gray-900">{commentAuthor}</p>
                    <p className="text-sm text-gray-700 break-words">{comment.content}</p>
                    <p className="text-xs text-gray-400">{timeAgo(comment.created_at)}</p>
                  </div>
                </div>
              );
            })
          ) : (
            <p className="text-sm text-gray-400 text-center py-2">No comments yet</p>
          )}
        </div>
      )}
      <form onSubmit={handleSubmit} className="mt-3 flex gap-2">
        <input
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Add a comment..."
          className="input-field flex-1 text-sm bg-gray-50 focus:bg-white"
        />
        <button
          type="submit"
          disabled={!text.trim() || addComment.isPending}
          className="btn-primary px-4 py-2 text-sm whitespace-nowrap"
        >
          {addComment.isPending ? '...' : 'Post'}
        </button>
      </form>
    </div>
  );
}