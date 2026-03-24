import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface AuthState {
  accessToken: string | null;
  userId: string | null;
  isAuthenticated: boolean;
  setAuth: (token: string, userId: string) => void;
  logout: () => void;
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set) => ({
      accessToken: null,
      userId: null,
      isAuthenticated: false,
      setAuth: (token, userId) =>
        set({ accessToken: token, userId, isAuthenticated: true }),
      logout: () =>
        set({ accessToken: null, userId: null, isAuthenticated: false }),
    }),
    {
      name: 'trueconnect-auth',
      partialize: (state) => ({
        accessToken: state.accessToken,
        userId: state.userId,
        isAuthenticated: state.isAuthenticated,
      }),
    }
  )
);
