import api from './api';
import { User } from '../types';

export interface LoginResponse {
  success: boolean;
  token: string;
  user: User;
}

export interface AuthError {
  success: boolean;
  error: string;
}

/**
 * Authenticate a user with username and password
 * @param username - The username
 * @param password - The password
 * @returns Promise with login response
 */
export const login = async (username: string, password: string): Promise<LoginResponse> => {
  try {
    const response = await api.post<LoginResponse>('/api/auth/login', { username, password });
    return response.data;
  } catch (error: any) {
    if (error.response && error.response.data) {
      throw error.response.data;
    }
    throw { success: false, error: 'Network error. Please try again.' };
  }
};

/**
 * Register a new user
 * @param userData - The user registration data
 * @returns Promise with login response
 */
export const register = async (userData: {
  username: string;
  email: string;
  password: string;
  firstName: string;
  lastName: string;
}): Promise<LoginResponse> => {
  try {
    const response = await api.post<LoginResponse>('/api/auth/register', userData);
    return response.data;
  } catch (error: any) {
    if (error.response && error.response.data) {
      throw error.response.data;
    }
    throw { success: false, error: 'Network error. Please try again.' };
  }
};

/**
 * Refresh the authentication token
 * @returns Promise with login response
 */
export const refreshToken = async (): Promise<LoginResponse> => {
  try {
    const response = await api.post<LoginResponse>('/api/auth/refresh');
    return response.data;
  } catch (error: any) {
    if (error.response && error.response.data) {
      throw error.response.data;
    }
    throw { success: false, error: 'Failed to refresh token' };
  }
};

/**
 * Get the current authenticated user
 * @returns Promise with user data
 */
export const getCurrentUser = async (): Promise<User> => {
  try {
    const response = await api.get<{ success: boolean; data: User }>('/api/users/me');
    return response.data.data;
  } catch (error: any) {
    if (error.response && error.response.data) {
      throw error.response.data;
    }
    throw { success: false, error: 'Failed to get current user' };
  }
};

/**
 * Log out the current user
 * @returns Promise<void>
 */
export const logout = async (): Promise<void> => {
  try {
    await api.post('/api/auth/logout');
  } catch (error) {
    console.error('Logout error:', error);
  }
};

/**
 * Check if user is authenticated
 * @returns boolean indicating authentication status
 */
export const isAuthenticated = (): boolean => {
  return localStorage.getItem('token') !== null;
};

/**
 * Get stored user data
 * @returns User object or null
 */
export const getStoredUser = (): User | null => {
  const userData = localStorage.getItem('user');
  if (userData) {
    try {
      return JSON.parse(userData);
    } catch (error) {
      console.error('Failed to parse user data:', error);
    }
  }
  return null;
};