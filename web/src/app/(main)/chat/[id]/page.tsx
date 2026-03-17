'use client';

import { useEffect, useRef, useState } from 'react';
import { useParams, useRouter } from 'next/navigation';
import { ArrowLeft, Send } from 'lucide-react';
import { useMessages, useMatches } from '@/hooks/api';
import { useChatStore } from '@/store/chat-store';
import { useAuthStore } from '@/store/auth-store';
import { wsService } from '@/lib/websocket';
import { LoadingScreen, Avatar } from '@/components/ui/common';
import { cn, timeAgo } from '@/lib/utils';
import type { Message } from '@/types';

export default function ChatPage() {
  const params = useParams<{ id: string }>();
  const router = useRouter();
  const matchId = params.id;
  const userId = useAuthStore((s) => s.userId);
  const { data: serverMessages, isLoading } = useMessages(matchId);
  const { data: matches } = useMatches();
  const { messages: storeMessages, addMessage, setMessages } = useChatStore();
  const [input, setInput] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const match = matches?.find((m) => m.id === matchId);
  const messages = storeMessages[matchId] || [];

  // Sync server messages to store
  useEffect(() => {
    if (serverMessages) {
      setMessages(matchId, serverMessages);
    }
  }, [serverMessages, matchId, setMessages]);

  // Connect WebSocket and listen for messages
  useEffect(() => {
    wsService.connect();
    const unsub = wsService.on('message', (data) => {
      const msg = data as unknown as Message;
      if (msg.match_id === matchId) {
        addMessage(matchId, msg);
      }
    });
    return () => {
      unsub();
    };
  }, [matchId, addMessage]);

  // Auto-scroll
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = () => {
    if (!input.trim()) return;
    wsService.sendMessage(matchId, input.trim());
    setInput('');
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  if (isLoading) return <LoadingScreen />;

  return (
    <div className="flex h-[calc(100vh-3rem)] flex-col md:h-[calc(100vh-1.5rem)]">
      {/* Header */}
      <div className="flex items-center gap-3 border-b border-gray-200 bg-white px-4 py-3 rounded-t-2xl">
        <button onClick={() => router.push('/matches')} className="text-gray-500 hover:text-gray-700 md:hidden">
          <ArrowLeft className="h-5 w-5" />
        </button>
        <Avatar
          src={match?.other_user?.avatar_url}
          name={match?.other_user?.display_name || 'User'}
          size="md"
        />
        <div>
          <h2 className="font-semibold text-gray-900">
            {match?.other_user?.display_name || 'Chat'}
          </h2>
        </div>
      </div>

      {/* Messages */}
      <div className="flex-1 overflow-y-auto bg-gray-50 p-4 space-y-3">
        {messages.length === 0 && (
          <p className="text-center text-sm text-gray-400 py-8">
            No messages yet. Say hello! 👋
          </p>
        )}
        {messages.map((msg) => {
          const isMe = msg.sender_id === userId;
          return (
            <div key={msg.id} className={cn('flex', isMe ? 'justify-end' : 'justify-start')}>
              <div
                className={cn(
                  'max-w-[70%] rounded-2xl px-4 py-2.5',
                  isMe
                    ? 'bg-primary-500 text-white rounded-br-md'
                    : 'bg-white text-gray-900 shadow-sm rounded-bl-md'
                )}
              >
                <p className="text-sm whitespace-pre-wrap">{msg.content}</p>
                <p
                  className={cn(
                    'mt-1 text-xs',
                    isMe ? 'text-white/60' : 'text-gray-400'
                  )}
                >
                  {timeAgo(msg.created_at)}
                </p>
              </div>
            </div>
          );
        })}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="flex items-center gap-2 border-t border-gray-200 bg-white p-4 rounded-b-2xl">
        <input
          value={input}
          onChange={(e) => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Type a message..."
          className="input-field flex-1"
        />
        <button
          onClick={handleSend}
          disabled={!input.trim()}
          className="flex h-11 w-11 items-center justify-center rounded-xl bg-primary-500 text-white transition-colors hover:bg-primary-600 disabled:opacity-50"
        >
          <Send className="h-5 w-5" />
        </button>
      </div>
    </div>
  );
}
