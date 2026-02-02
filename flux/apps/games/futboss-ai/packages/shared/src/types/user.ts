// FutBoss AI - User Types
// Author: Bruno Lucena (bruno@lucena.cloud)

export interface User {
  id: string;
  username: string;
  email: string;
  tokens: number;
  createdAt: Date;
  isActive: boolean;
}

export interface UserCreate {
  username: string;
  email: string;
  password: string;
}

export interface UserLogin {
  email: string;
  password: string;
}

export interface UserResponse {
  id: string;
  username: string;
  email: string;
  tokens: number;
  createdAt: Date;
}

export interface AuthResponse {
  accessToken: string;
  tokenType: string;
  user: UserResponse;
}

