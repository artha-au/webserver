import React, { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { api, tokenManager } from '../services/api';
import toast from 'react-hot-toast';

interface User {
  id: string;
  email: string;
  name: string;
  roles?: string[];
  teams?: Array<{
    id: string;
    name: string;
    role: string;
  }>;
}

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  refreshUser: () => Promise<void>;
  isAdmin: boolean;
  isTeamLeader: boolean;
}

const AuthContext = createContext<AuthContextType | null>(null);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  // Check if user has admin role
  const isAdmin = user?.roles?.includes('admin') || 
                  user?.roles?.includes('super_admin') || false;
  
  // Check if user is a team leader in any team
  const isTeamLeader = user?.teams?.some(team => team.role === 'leader') || false;

  // Load user from token on mount
  useEffect(() => {
    const loadUser = async () => {
      const token = tokenManager.getToken();
      if (token) {
        try {
          const response = await api.auth.getProfile();
          setUser(response.data);
        } catch (error) {
          console.error('Failed to load user profile:', error);
          tokenManager.clearTokens();
        }
      }
      setLoading(false);
    };
    loadUser();
  }, []);

  const login = async (email: string, password: string) => {
    try {
      const response = await api.auth.login(email, password);
      const { accessToken, refreshToken, user: userData } = response.data;
      
      tokenManager.setToken(accessToken);
      if (refreshToken) {
        tokenManager.setRefreshToken(refreshToken);
      }
      
      setUser(userData);
      toast.success('Successfully logged in!');
    } catch (error: any) {
      console.error('Login failed:', error);
      throw new Error(error.response?.data?.error || 'Login failed');
    }
  };

  const logout = () => {
    tokenManager.clearTokens();
    setUser(null);
    toast.success('Successfully logged out');
    window.location.href = '/login';
  };

  const refreshUser = async () => {
    try {
      const response = await api.auth.getProfile();
      setUser(response.data);
    } catch (error) {
      console.error('Failed to refresh user profile:', error);
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        loading,
        login,
        logout,
        refreshUser,
        isAdmin,
        isTeamLeader,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};