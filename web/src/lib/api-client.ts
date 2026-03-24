import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { useAuthStore } from '@/store/auth-store';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

const PUBLIC_ENDPOINTS = ['/v1/auth/login', '/v1/auth/register', '/v1/auth/refresh', '/v1/health'];

export const api = axios.create({
  baseURL: API_BASE,
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
  withCredentials: true,
});

// Request interceptor: attach JWT
api.interceptors.request.use((config: InternalAxiosRequestConfig) => {
  const isPublic = PUBLIC_ENDPOINTS.some((ep) => config.url?.includes(ep));
  if (!isPublic) {
    const token = useAuthStore.getState().accessToken;
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
  }
  return config;
});

// Response interceptor: refresh on 401
let isRefreshing = false;
let failedQueue: Array<{
  resolve: (value: unknown) => void;
  reject: (reason: unknown) => void;
}> = [];

function processQueue(error: unknown, token: string | null = null) {
  failedQueue.forEach((prom) => {
    if (error) {
      prom.reject(error);
    } else {
      prom.resolve(token);
    }
  });
  failedQueue = [];
}

api.interceptors.response.use(
  (response) => {
    // If the response contains a 'meta' object (e.g., pagination), preserve it
    // by attaching it to the data array/object, or returning the entire envelope.
    // For our hooks, if there's meta, we'll return the whole envelope as expected.
    if (response.data && response.data.data !== undefined) {
      if (response.data.meta !== undefined) {
          // If there's metadata, keep the envelope so hooks can read items and cursors
          response.data = { items: response.data.data, ...response.data.meta };
      } else {
          // Flatten data directly if no meta is present
          response.data = response.data.data;
      }
    }
    return response;
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        return new Promise((resolve, reject) => {
          failedQueue.push({ resolve, reject });
        }).then((token) => {
          originalRequest.headers.Authorization = `Bearer ${token}`;
          return api(originalRequest);
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        const { data } = await axios.post(
          `${API_BASE}/v1/auth/refresh`,
          {},
          { withCredentials: true }
        );
        const backendData = data.data !== undefined ? data.data : data;
        const newToken = backendData.access_token;
        useAuthStore.getState().setAuth(newToken, backendData.user_id);
        processQueue(null, newToken);
        originalRequest.headers.Authorization = `Bearer ${newToken}`;
        return api(originalRequest);
      } catch (refreshError) {
        processQueue(refreshError, null);
        useAuthStore.getState().logout();
        if (typeof window !== 'undefined') {
          window.location.href = '/login';
        }
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  }
);

export default api;
