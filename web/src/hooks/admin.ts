import { useQuery } from '@tanstack/react-query';
import api from '@/lib/api-client';
import { SybilCluster } from '@/types';

export function useSybilClusters() {
  return useQuery({
    queryKey: ['sybilClusters'],
    queryFn: async () => {
      const res = await api.get<SybilCluster[]>('/v1/admin/sybil-clusters');
      return res.data;
    }
  });
}