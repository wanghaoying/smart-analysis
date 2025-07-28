import api from './api';
import { 
  LoginRequest, 
  RegisterRequest, 
  LoginResponse, 
  User, 
  ApiResponse 
} from './types';

export const authService = {
  // 用户登录
  async login(data: LoginRequest): Promise<LoginResponse> {
    const response: ApiResponse<LoginResponse> = await api.post('/user/login', data);
    if (response.data) {
      localStorage.setItem('token', response.data.token);
      localStorage.setItem('user', JSON.stringify(response.data.user));
    }
    return response.data!;
  },

  // 用户注册
  async register(data: RegisterRequest): Promise<LoginResponse> {
    const response: ApiResponse<LoginResponse> = await api.post('/user/register', data);
    if (response.data) {
      localStorage.setItem('token', response.data.token);
      localStorage.setItem('user', JSON.stringify(response.data.user));
    }
    return response.data!;
  },

  // 获取用户资料
  async getProfile(): Promise<User> {
    const response: ApiResponse<User> = await api.get('/user/profile');
    return response.data!;
  },

  // 更新用户资料
  async updateProfile(data: Partial<User>): Promise<User> {
    const response: ApiResponse<User> = await api.put('/user/profile', data);
    return response.data!;
  },

  // 退出登录
  logout(): void {
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    window.location.href = '/login';
  },

  // 获取当前用户
  getCurrentUser(): User | null {
    const userStr = localStorage.getItem('user');
    return userStr ? JSON.parse(userStr) : null;
  },

  // 检查是否已登录
  isAuthenticated(): boolean {
    return !!localStorage.getItem('token');
  },
};
