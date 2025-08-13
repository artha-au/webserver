import axios, { AxiosError, AxiosInstance, AxiosRequestConfig } from 'axios';
import toast from 'react-hot-toast';

// API configuration
const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// Create axios instance
const apiClient: AxiosInstance = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
  timeout: 30000,
});

// Token management
export const tokenManager = {
  getToken: (): string | null => localStorage.getItem('accessToken'),
  setToken: (token: string) => localStorage.setItem('accessToken', token),
  getRefreshToken: (): string | null => localStorage.getItem('refreshToken'),
  setRefreshToken: (token: string) => localStorage.setItem('refreshToken', token),
  clearTokens: () => {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
  },
};

// Request interceptor to add auth token
apiClient.interceptors.request.use(
  (config) => {
    const token = tokenManager.getToken();
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Response interceptor for error handling
apiClient.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const originalRequest = error.config as AxiosRequestConfig & { _retry?: boolean };
    
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;
      
      try {
        const refreshToken = tokenManager.getRefreshToken();
        if (refreshToken) {
          const response = await axios.post(`${API_BASE_URL}/auth/refresh`, {
            refreshToken,
          });
          
          const { accessToken } = response.data;
          tokenManager.setToken(accessToken);
          
          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${accessToken}`;
          }
          
          return apiClient(originalRequest);
        }
      } catch (refreshError) {
        tokenManager.clearTokens();
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }
    
    // Show error toast for user-facing errors
    if (error.response?.data && typeof error.response.data === 'object' && 'error' in error.response.data) {
      const errorMessage = (error.response.data as { error: string }).error;
      toast.error(errorMessage);
    } else if (error.message) {
      toast.error(error.message);
    }
    
    return Promise.reject(error);
  }
);

// API service methods
export const api = {
  // Auth endpoints
  auth: {
    login: (email: string, password: string) =>
      apiClient.post('/auth/token', { email, password }),
    logout: () => apiClient.post('/auth/logout'),
    refresh: (refreshToken: string) =>
      apiClient.post('/auth/refresh', { refreshToken }),
    getProfile: () => apiClient.get('/auth/profile'),
  },

  // Admin team endpoints
  adminTeams: {
    list: (limit = 50, offset = 0) =>
      apiClient.get('/api/v1/admin/teams', { params: { limit, offset } }),
    create: (data: any) => apiClient.post('/api/v1/admin/teams', data),
    get: (teamId: string) => apiClient.get(`/api/v1/admin/teams/${teamId}`),
    update: (teamId: string, data: any) =>
      apiClient.put(`/api/v1/admin/teams/${teamId}`, data),
    delete: (teamId: string) => apiClient.delete(`/api/v1/admin/teams/${teamId}`),
  },

  // Admin team member endpoints
  adminMembers: {
    list: (teamId: string) =>
      apiClient.get(`/api/v1/admin/teams/${teamId}/members`),
    add: (teamId: string, data: any) =>
      apiClient.post(`/api/v1/admin/teams/${teamId}/members`, data),
    get: (teamId: string, memberId: string) =>
      apiClient.get(`/api/v1/admin/teams/${teamId}/members/${memberId}`),
    update: (teamId: string, memberId: string, data: any) =>
      apiClient.put(`/api/v1/admin/teams/${teamId}/members/${memberId}`, data),
    remove: (teamId: string, memberId: string) =>
      apiClient.delete(`/api/v1/admin/teams/${teamId}/members/${memberId}`),
  },

  // User team endpoints
  teams: {
    listMyTeams: () => apiClient.get('/api/v1/teams'),
    getMembers: (teamId: string) =>
      apiClient.get(`/api/v1/teams/${teamId}/members`),
  },

  // Timesheet endpoints
  timesheets: {
    listMy: (teamId: string, limit = 50, offset = 0) =>
      apiClient.get(`/api/v1/teams/${teamId}/timesheets`, {
        params: { limit, offset },
      }),
    create: (teamId: string, data: any) =>
      apiClient.post(`/api/v1/teams/${teamId}/timesheets`, data),
    get: (teamId: string, timesheetId: string) =>
      apiClient.get(`/api/v1/teams/${teamId}/timesheets/${timesheetId}`),
    update: (teamId: string, timesheetId: string, data: any) =>
      apiClient.put(`/api/v1/teams/${teamId}/timesheets/${timesheetId}`, data),
    delete: (teamId: string, timesheetId: string) =>
      apiClient.delete(`/api/v1/teams/${teamId}/timesheets/${timesheetId}`),
    submit: (teamId: string, timesheetId: string) =>
      apiClient.post(`/api/v1/teams/${teamId}/timesheets/${timesheetId}/submit`),
    
    // Team leader endpoints
    listMemberTimesheets: (teamId: string, memberId: string, limit = 50, offset = 0) =>
      apiClient.get(`/api/v1/teams/${teamId}/members/${memberId}/timesheets`, {
        params: { limit, offset },
      }),
    approve: (teamId: string, memberId: string, timesheetId: string) =>
      apiClient.post(
        `/api/v1/teams/${teamId}/members/${memberId}/timesheets/${timesheetId}/approve`
      ),
    reject: (teamId: string, memberId: string, timesheetId: string, reason: string) =>
      apiClient.post(
        `/api/v1/teams/${teamId}/members/${memberId}/timesheets/${timesheetId}/reject`,
        { reason }
      ),
  },

  // Roster endpoints
  rosters: {
    list: (teamId: string, limit = 50, offset = 0) =>
      apiClient.get(`/api/v1/teams/${teamId}/rosters`, {
        params: { limit, offset },
      }),
    create: (teamId: string, data: any) =>
      apiClient.post(`/api/v1/teams/${teamId}/rosters`, data),
    get: (teamId: string, rosterId: string) =>
      apiClient.get(`/api/v1/teams/${teamId}/rosters/${rosterId}`),
    update: (teamId: string, rosterId: string, data: any) =>
      apiClient.put(`/api/v1/teams/${teamId}/rosters/${rosterId}`, data),
    delete: (teamId: string, rosterId: string) =>
      apiClient.delete(`/api/v1/teams/${teamId}/rosters/${rosterId}`),
  },
};

export default apiClient;