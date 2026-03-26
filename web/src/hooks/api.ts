import { useMutation, useQuery, useQueryClient, useInfiniteQuery } from '@tanstack/react-query';
import api from '@/lib/api-client';
import { useAuthStore } from '@/store/auth-store';
import type {
  AuthResponse,
  User,
  Profile,
  ProfileUpsert,
  ProfilePhoto,
  Match,
  MatchCandidate,
  Message,
  Post,
  PostComment,
  UserSettings,
  TrustScore,
  KycStatus,
  PaginatedResponse,
} from '@/types';

// ─── Auth ───────────────────────────────────────────────

export function useLogin() {
  const setAuth = useAuthStore((s) => s.setAuth);
  return useMutation({
    mutationFn: async (data: { phone: string; password: string }) => {
      const res = await api.post<AuthResponse>('/v1/auth/login', data);
      return res.data;
    },
    onSuccess: (data) => setAuth(data.access_token, data.user_id),
  });
}

export function useRegister() {
  const setAuth = useAuthStore((s) => s.setAuth);
  return useMutation({
    mutationFn: async (data: { phone: string; password: string }) => {
      const res = await api.post<AuthResponse>('/v1/auth/register', data);
      return res.data;
    },
    onSuccess: (data) => setAuth(data.access_token, data.user_id),
  });
}

export function useLogout() {
  const logout = useAuthStore((s) => s.logout);
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async () => {
      await api.post('/v1/auth/logout');
    },
    onSettled: () => {
      logout();
      queryClient.clear();
    },
  });
}

// ─── User ───────────────────────────────────────────────

export function useCurrentUser() {
  return useQuery({
    queryKey: ['currentUser'],
    queryFn: async () => {
      const res = await api.get<User>('/v1/users/me');
      return res.data;
    },
  });
}

// ─── Profile ────────────────────────────────────────────

export function useMyProfile() {
  return useQuery({
    queryKey: ['myProfile'],
    queryFn: async () => {
      const res = await api.get<Profile>('/v1/profiles/me');
      return res.data;
    },
    retry: false,
  });
}

export function useProfile(userId: string) {
  return useQuery({
    queryKey: ['profile', userId],
    queryFn: async () => {
      const res = await api.get<Profile>(`/v1/profiles/${userId}`);
      return res.data;
    },
    enabled: !!userId,
  });
}

export function useUpdateProfile() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (data: ProfileUpsert) => {
      const res = await api.put<Profile>('/v1/profiles/me', data);
      return res.data;
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['myProfile'] }),
  });
}

export function useMyPhotos() {
  return useQuery({
    queryKey: ['myPhotos'],
    queryFn: async () => {
      const res = await api.get<ProfilePhoto[]>('/v1/profiles/me/photos');
      return res.data;
    },
  });
}

export function useUploadPhoto() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (file: File) => {
      const form = new FormData();
      form.append('photo', file);
      const res = await api.post<ProfilePhoto>('/v1/profiles/me/photos', form, {
      });
      return res.data;
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['myPhotos'] }),
  });
}

export function useDeletePhoto() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (photoId: string) => {
      await api.delete(`/v1/profiles/me/photos/${photoId}`);
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['myPhotos'] }),
  });
}

// ─── Matching ───────────────────────────────────────────

export function useCandidates() {
  return useInfiniteQuery({
    queryKey: ['candidates'],
    queryFn: async ({ pageParam = 1 }) => {
      const res = await api.get<PaginatedResponse<MatchCandidate>>('/v1/matching/candidates', {
        params: { page: pageParam, page_size: 10 },
      });
      // ДОБАВЬ ЭТУ СТРОЧКУ:
      console.log("🔥 ОТВЕТ ОТ БЭКЕНДА (CANDIDATES):", res.data);
      return res.data;
    },
    getNextPageParam: (last, pages) => (last.has_more ? pages.length + 1 : undefined),
    initialPageParam: 1,
  });
}
export function useLikeUser() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (targetId: string) => {
      const res = await api.post<{ matched: boolean }>('/v1/matching/like', { target_id: targetId });
      return res.data;
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['matches'] }),
  });
}

export function usePassUser() {
  return useMutation({
    mutationFn: async (targetId: string) => {
      await api.post('/v1/matching/pass', { target_id: targetId });
    },
  });
}

export function useMatches() {
  return useQuery({
    queryKey: ['matches'],
    queryFn: async () => {
      // Мы используем тип PaginatedResponse, так как бэкенд присылает 'items'
      const res = await api.get<PaginatedResponse<Match>>('/v1/matches');
      
      console.log("🍏 Тело ответа (body):", res.data);
      
      // ВОТ ОНО РЕШЕНИЕ: Достаем массив из поля items
      const matchesArray = res.data.items; 
      
      console.log("🚀 Итоговый массив мэтчей:", matchesArray);

      if (!Array.isArray(matchesArray)) {
        console.error('❌ ОШИБКА: Поле items не является массивом!', res.data);
        return [];
      }

      return matchesArray;
    },
  });
}
// ─── Chat ───────────────────────────────────────────────

export function useMessages(matchId: string) {
  return useInfiniteQuery({
    queryKey: ['messages', matchId],
    queryFn: async ({ pageParam = '' }) => {
      const res = await api.get<PaginatedResponse<Message>>(`/v1/matches/${matchId}/messages`, {
        params: { cursor: pageParam, limit: 20 },
      });
      return res.data;
    },
    getNextPageParam: (last) => last.next_cursor || undefined,
    initialPageParam: '',
    enabled: !!matchId,
  });
}

// ─── Feed ───────────────────────────────────────────────

export function usePosts() {
  return useInfiniteQuery({
    queryKey: ['posts'],
    queryFn: async ({ pageParam = '' }) => {
      const res = await api.get<PaginatedResponse<Post>>('/v1/posts', {
        params: { cursor: pageParam, limit: 20 },
      });
      return res.data;
    },
    getNextPageParam: (last) => last.next_cursor || undefined,
    initialPageParam: '',
  });
}

export function useCreatePost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (data: { content: string; media?: File }) => {
      if (data.media) {
        const form = new FormData();
        form.append('content', data.content);
        form.append('media', data.media);
        
        // ИСПРАВЛЕНО: Убрали headers, чтобы браузер сам правильно собрал multipart/form-data
        const res = await api.post<Post>('/v1/posts', form);
        return res.data;
      }
      const res = await api.post<Post>('/v1/posts', { content: data.content });
      return res.data;
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['posts'] }),
  });
}
export function useDeletePost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (postId: string) => {
      await api.delete(`/v1/posts/${postId}`);
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['posts'] }),
  });
}

export function useLikePost() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async ({ postId, liked }: { postId: string; liked: boolean }) => {
      if (liked) {
        await api.delete(`/v1/posts/${postId}/like`);
      } else {
        await api.post(`/v1/posts/${postId}/like`);
      }
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['posts'] }),
  });
}

export function useComments(postId: string) {
  return useQuery({
    queryKey: ['comments', postId],
    queryFn: async () => {
      const res = await api.get<PostComment[]>(`/v1/posts/${postId}/comments`);
      return res.data;
    },
    enabled: !!postId,
  });
}

export function useAddComment() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (data: { postId: string; content: string }) => {
      const res = await api.post<PostComment>(`/v1/posts/${data.postId}/comments`, {
        content: data.content,
      });
      return res.data;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['comments', variables.postId] });
      queryClient.invalidateQueries({ queryKey: ['posts'] });
    },
  });
}

// ─── Settings ───────────────────────────────────────────

export function useSettings() {
  return useQuery({
    queryKey: ['settings'],
    queryFn: async () => {
      const res = await api.get<UserSettings>('/v1/settings');
      return res.data;
    },
  });
}

export function useUpdateSettings() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (data: Partial<UserSettings>) => {
      const res = await api.patch<UserSettings>('/v1/settings', data);
      return res.data;
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['settings'] }),
  });
}

// ─── Trust / Reputation ─────────────────────────────────

export function useReputation(userId: string) {
  return useQuery({
    queryKey: ['reputation', userId],
    queryFn: async () => {
      const res = await api.get<TrustScore>(`/v1/users/${userId}/reputation`);
      return res.data;
    },
    enabled: !!userId,
  });
}

// ─── KYC ────────────────────────────────────────────────

export function useKycStatus() {
  return useQuery({
    queryKey: ['kycStatus'],
    queryFn: async () => {
      const res = await api.get<KycStatus>('/v1/kyc/status');
      return res.data;
    },
  });
}

export function useSubmitKyc() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (file: File) => {
      const form = new FormData();
      form.append('document', file);
      await api.post('/v1/kyc/submit', form, {
        headers: { 'Content-Type': 'multipart/form-data' },
      });
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['kycStatus'] }),
  });
}

// ─── Reports ────────────────────────────────────────────

export function useReportUser() {
  return useMutation({
    mutationFn: async (data: { reported_id: string; reason: string }) => {
      await api.post('/v1/reports', data);
    },
  });
}
