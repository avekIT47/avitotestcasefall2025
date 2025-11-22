import axios, { type AxiosInstance } from 'axios';
import type {
  Team,
  User,
  PullRequest,
  Statistics,
  CreateTeamRequest,
  CreateUserRequest,
  UpdateUserRequest,
  CreatePullRequestRequest,
  AddUserToTeamRequest,
  RemoveUserFromTeamRequest,
  DeactivateUsersRequest,
  ReassignReviewerRequest,
} from '../types';

class ApiService {
  private api: AxiosInstance;

  constructor() {
    // В dev режиме используем /api префикс (Vite proxy)
    // В production nginx будет проксировать /api на backend
    const baseURL = import.meta.env.DEV 
      ? '/api'  // Для dev режима - через Vite proxy
      : (import.meta.env.VITE_API_URL || '/api'); // Для production - через nginx
    
    this.api = axios.create({
      baseURL,
      headers: {
        'Content-Type': 'application/json',
      },
    });

    // Request interceptor
    this.api.interceptors.request.use(
      (config) => {
        // Можно добавить токен авторизации здесь
        const token = localStorage.getItem('authToken');
        if (token) {
          config.headers.Authorization = `Bearer ${token}`;
        }
        return config;
      },
      (error) => {
        return Promise.reject(error);
      }
    );

    // Response interceptor
    this.api.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          // Обработка неавторизованного доступа
          localStorage.removeItem('authToken');
          window.location.href = '/login';
        }
        return Promise.reject(error);
      }
    );
  }

  // Health check
  async health(): Promise<{ status: string }> {
    const response = await this.api.get('/health');
    return response.data;
  }

  // Teams
  async getTeams(): Promise<Team[]> {
    const response = await this.api.get('/teams');
    return response.data;
  }

  async getTeam(teamId: number): Promise<Team> {
    const response = await this.api.get(`/teams/${teamId}`);
    return response.data;
  }

  async createTeam(data: CreateTeamRequest): Promise<Team> {
    const response = await this.api.post('/teams', data);
    return response.data;
  }

  async deleteTeam(teamId: number): Promise<void> {
    await this.api.delete(`/teams/${teamId}`);
  }

  async addUserToTeam(teamId: number, data: AddUserToTeamRequest): Promise<void> {
    await this.api.post(`/teams/${teamId}/users`, data);
  }

  async removeUserFromTeam(teamId: number, data: RemoveUserFromTeamRequest): Promise<void> {
    await this.api.delete(`/teams/${teamId}/users`, { data });
  }

  async deactivateUsers(teamId: number, data: DeactivateUsersRequest): Promise<void> {
    await this.api.post(`/teams/${teamId}/users/deactivate`, data);
  }

  // Users
  async getUsers(params?: { teamId?: number; isActive?: boolean }): Promise<User[]> {
    const response = await this.api.get('/users', { params });
    return response.data;
  }

  async getUser(userId: number): Promise<User> {
    const response = await this.api.get(`/users/${userId}`);
    return response.data;
  }

  async createUser(data: CreateUserRequest): Promise<User> {
    const response = await this.api.post('/users', data);
    return response.data;
  }

  async updateUser(userId: number, data: UpdateUserRequest): Promise<User> {
    const response = await this.api.patch(`/users/${userId}`, data);
    return response.data;
  }

  // Pull Requests
  async getPullRequests(params?: { 
    authorId?: number; 
    teamId?: number; 
    status?: string;
    reviewerId?: number;
  }): Promise<PullRequest[]> {
    const response = await this.api.get('/pull-requests', { params });
    return response.data;
  }

  async getPullRequest(prId: number): Promise<PullRequest> {
    const response = await this.api.get(`/pull-requests/${prId}`);
    return response.data;
  }

  async createPullRequest(data: CreatePullRequestRequest): Promise<PullRequest> {
    const response = await this.api.post('/pull-requests', data);
    return response.data;
  }

  async addReviewer(prId: number, reviewerId: number): Promise<PullRequest> {
    const response = await this.api.post(`/pull-requests/${prId}/reviewers`, { reviewerId });
    return response.data;
  }

  async reassignReviewer(prId: number, data: ReassignReviewerRequest): Promise<PullRequest> {
    const response = await this.api.put(`/pull-requests/${prId}/reviewers`, data);
    return response.data;
  }

  async mergePullRequest(prId: number): Promise<PullRequest> {
    const response = await this.api.post(`/pull-requests/${prId}/merge`);
    return response.data;
  }

  async closePullRequest(prId: number): Promise<PullRequest> {
    const response = await this.api.post(`/pull-requests/${prId}/close`);
    return response.data;
  }

  // Statistics
  async getStatistics(): Promise<Statistics> {
    const response = await this.api.get('/statistics');
    return response.data;
  }
}

export const apiService = new ApiService();
