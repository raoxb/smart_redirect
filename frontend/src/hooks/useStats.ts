import { useQuery } from '@tanstack/react-query';
import { api } from '../services/api';

export interface RealtimeStats {
  hourly: Array<{
    hour: string;
    visits: number;
    unique_ips: number;
    redirects: number;
    blocked: number;
  }>;
  geographic: Array<{
    country_code: string;
    country_name: string;
    count: number;
    percentage: number;
  }>;
  top_targets: Array<{
    target_id: number;
    url: string;
    hits: number;
    percentage: number;
  }>;
  summary: {
    total_links: number;
    active_links: number;
    today_visits: number;
    week_visits: number;
    avg_response_time: string;
    success_rate: string;
  };
  last_updated: string;
}

export const useRealtimeStats = (hours: number = 24) => {
  return useQuery<RealtimeStats>({
    queryKey: ['stats', 'realtime', hours],
    queryFn: async () => {
      const response = await api.get(`/stats/realtime?hours=${hours}`);
      return response.data;
    },
    refetchInterval: 30000, // Refresh every 30 seconds
  });
};

export const useLinkStats = (linkId: string) => {
  return useQuery({
    queryKey: ['stats', 'link', linkId],
    queryFn: async () => {
      const response = await api.get(`/stats/links/${linkId}`);
      return response.data;
    },
    enabled: !!linkId,
  });
};

export const useTargetStats = (targetId: string) => {
  return useQuery({
    queryKey: ['stats', 'target', targetId],
    queryFn: async () => {
      const response = await api.get(`/stats/targets/${targetId}`);
      return response.data;
    },
    enabled: !!targetId,
  });
};