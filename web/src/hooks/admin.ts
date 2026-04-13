import { useQuery } from '@tanstack/react-query';
import api from '@/lib/api-client';
import { SybilCluster, AdminDashboardStats, AdminUserRow } from '@/types';

export function useSybilClusters() {
  return useQuery({
    queryKey: ['sybilClusters'],
    queryFn: async () => {
      const res = await api.get<SybilCluster[]>('/v1/admin/sybil-clusters');
      return res.data;
    }
  });
}

export function useAdminSystemStats() {
  return useQuery({
    queryKey: ['adminDashboardStats'],
    queryFn: async () => {
      const res = await api.get<AdminDashboardStats>('/v1/admin/analytics');
      return res.data;
    }
  });
}

export function useAdminSearchUsers(q: string, limit: number = 20, offset: number = 0) {
  return useQuery({
    queryKey: ['adminSearchUsers', q, limit, offset],
    queryFn: async () => {
      const res = await api.get<{
        items: AdminUserRow[];
        total: number;
      }>('/v1/admin/users', { params: { q, limit, offset } });
      return res.data;
    }
  });
}