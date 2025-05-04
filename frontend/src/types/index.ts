export interface User {
  id: string;
  username: string;
  email: string;
  firstName: string;
  lastName: string;
  active: boolean;
  emailVerified: boolean;
  roleIds: string[];
  organizationIds: string[];
  authProvider: string;
  createdAt: string;
  updatedAt: string;
  lastLogin?: string;
}

export interface Organization {
  id: string;
  name: string;
  description: string;
  domain?: string;
  active: boolean;
  adminIds: string[];
  createdAt: string;
  updatedAt: string;
}

export interface Role {
  id: string;
  name: string;
  description: string;
  organizationId?: string;
  permissionIds: string[];
  isSystemDefault: boolean;
  createdAt: string;
  updatedAt: string;
}

export interface Permission {
  id: string;
  name: string;
  description: string;
  resource: string;
  action: string;
  organizationId?: string;
  isSystemDefault?: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface ApiResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginatedResponse<T> {
  success: boolean;
  data: {
    total: number;
    limit: number;
    skip: number;
    [key: string]: any; // This allows for different data properties like 'users', 'roles', etc.
  };
  error?: string;
}