'use client';

import { useState, useRef } from 'react';
import { X, ImagePlus } from 'lucide-react';
import { useCreatePost } from '@/hooks/api';
import { Spinner } from '@/components/ui/common';
import toast from 'react-hot-toast';

export function CreatePostModal({ onClose }: { onClose: () => void }) {
  const [content, setContent] = useState('');
  const [media, setMedia] = useState<File | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const fileRef = useRef<HTMLInputElement>(null);
  const createPost = useCreatePost();

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) {
      setMedia(file);
      setPreview(URL.createObjectURL(file));
    }
  };

  const handleSubmit = () => {
    if (!content.trim()) return;
    createPost.mutate(
      { content: content.trim(), media: media || undefined },
      {
        onSuccess: () => {
          toast.success('Post created!');
          onClose();
        },
        onError: () => toast.error('Failed to create post'),
      }
    );
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4">
      <div className="w-full max-w-lg rounded-2xl bg-white">
        {/* Header */}
        <div className="flex items-center justify-between border-b border-gray-200 p-4">
          <h2 className="text-lg font-semibold text-gray-900">Create Post</h2>
          <button onClick={onClose} className="rounded-lg p-1.5 text-gray-400 hover:bg-gray-100">
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Body */}
        <div className="p-4">
          <textarea
            value={content}
            onChange={(e) => setContent(e.target.value)}
            placeholder="What's on your mind?"
            rows={4}
            maxLength={1000}
            className="input-field resize-none"
          />
          <p className="mt-1 text-right text-xs text-gray-400">{content.length}/1000</p>

          {preview && (
            <div className="relative mt-3">
              <img src={preview} alt="Preview" className="w-full rounded-xl max-h-60 object-cover" />
              <button
                onClick={() => {
                  setMedia(null);
                  setPreview(null);
                }}
                className="absolute right-2 top-2 rounded-full bg-black/50 p-1.5 text-white hover:bg-black/70"
              >
                <X className="h-4 w-4" />
              </button>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between border-t border-gray-200 p-4">
          <button
            onClick={() => fileRef.current?.click()}
            className="flex items-center gap-2 rounded-lg px-3 py-2 text-sm text-gray-600 hover:bg-gray-100"
          >
            <ImagePlus className="h-5 w-5" />
            Photo
          </button>
          <input ref={fileRef} type="file" accept="image/*" className="hidden" onChange={handleFileChange} />
          <button
            onClick={handleSubmit}
            disabled={!content.trim() || createPost.isPending}
            className="btn-primary"
          >
            {createPost.isPending ? <Spinner className="h-5 w-5 text-white" /> : 'Publish'}
          </button>
        </div>
      </div>
    </div>
  );
}
