import { create } from 'zustand';
import { Message } from '@/types';

interface ChatState {
  messages: Record<string, Message[]>;
  activeMatchId: string | null;
  setActiveMatch: (matchId: string | null) => void;
  addMessage: (matchId: string, message: Message) => void;
  setMessages: (matchId: string, messages: Message[]) => void;
  clearMessages: (matchId: string) => void;
}

export const useChatStore = create<ChatState>()((set) => ({
  messages: {},
  activeMatchId: null,
  setActiveMatch: (matchId) => set({ activeMatchId: matchId }),
  addMessage: (matchId, message) =>
    set((state) => ({
      messages: {
        ...state.messages,
        [matchId]: [...(state.messages[matchId] || []), message],
      },
    })),
  setMessages: (matchId, messages) =>
    set((state) => ({
      messages: { ...state.messages, [matchId]: messages },
    })),
  clearMessages: (matchId) =>
    set((state) => {
      const newMessages = { ...state.messages };
      delete newMessages[matchId];
      return { messages: newMessages };
    }),
}));
