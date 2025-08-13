import React, { useState } from 'react';
import {
  Box,
  Card,
  TextField,
  Button,
  Typography,
  Alert,
  CircularProgress,
  InputAdornment,
  IconButton,
  Link,
  Divider,
} from '@mui/material';
import {
  Visibility,
  VisibilityOff,
  Email as EmailIcon,
  Lock as LockIcon,
} from '@mui/icons-material';
import { useNavigate } from 'react-router-dom';
import { useForm } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import { motion } from 'framer-motion';
import { useAuth } from '../../contexts/AuthContext';

const schema = yup.object({
  email: yup.string().email('Invalid email').required('Email is required'),
  password: yup.string().min(1, 'Password must be at least 6 characters').required('Password is required'),
});

interface LoginFormData {
  email: string;
  password: string;
}

const Login: React.FC = () => {
  const navigate = useNavigate();
  const { login } = useAuth();
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginFormData>({
    resolver: yupResolver(schema),
  });

  const onSubmit = async (data: LoginFormData) => {
    setLoading(true);
    setError(null);
    try {
      await login(data.email, data.password);
      navigate('/dashboard');
    } catch (err: any) {
      setError(err.message || 'Login failed. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  return (
    <Box
      sx={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        padding: 2,
      }}
    >
      <motion.div
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        transition={{ duration: 0.3 }}
      >
        <Card
          sx={{
            maxWidth: 450,
            width: '100%',
            p: 4,
            borderRadius: 3,
            boxShadow: '0 20px 60px rgba(0,0,0,0.3)',
          }}
        >
          {/* Logo and Title */}
          <Box sx={{ textAlign: 'center', mb: 4 }}>
            <Box
              sx={{
                width: 80,
                height: 80,
                borderRadius: 3,
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                margin: '0 auto',
                mb: 2,
                boxShadow: '0 10px 30px rgba(102, 126, 234, 0.4)',
              }}
            >
              <Typography
                variant="h3"
                sx={{ color: 'white', fontWeight: 'bold' }}
              >
                CRM
              </Typography>
            </Box>
            <Typography variant="h5" sx={{ fontWeight: 600, mb: 1 }}>
              Welcome Back
            </Typography>
            <Typography variant="body2" color="text.secondary">
              Sign in to continue to your workspace
            </Typography>
          </Box>

          {/* Error Alert */}
          {error && (
            <Alert severity="error" sx={{ mb: 3 }}>
              {error}
            </Alert>
          )}

          {/* Login Form */}
          <form onSubmit={handleSubmit(onSubmit)}>
            <TextField
              {...register('email')}
              fullWidth
              label="Email Address"
              placeholder="name@company.com"
              error={!!errors.email}
              helperText={errors.email?.message}
              sx={{ mb: 2 }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <EmailIcon color="action" />
                  </InputAdornment>
                ),
              }}
            />

            <TextField
              {...register('password')}
              fullWidth
              type={showPassword ? 'text' : 'password'}
              label="Password"
              placeholder="Enter your password"
              error={!!errors.password}
              helperText={errors.password?.message}
              sx={{ mb: 3 }}
              InputProps={{
                startAdornment: (
                  <InputAdornment position="start">
                    <LockIcon color="action" />
                  </InputAdornment>
                ),
                endAdornment: (
                  <InputAdornment position="end">
                    <IconButton
                      onClick={() => setShowPassword(!showPassword)}
                      edge="end"
                      size="small"
                    >
                      {showPassword ? <VisibilityOff /> : <Visibility />}
                    </IconButton>
                  </InputAdornment>
                ),
              }}
            />

            <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 3 }}>
              <Link
                href="#"
                underline="hover"
                sx={{ fontSize: '0.875rem' }}
              >
                Forgot password?
              </Link>
            </Box>

            <Button
              type="submit"
              fullWidth
              variant="contained"
              size="large"
              disabled={loading}
              sx={{
                py: 1.5,
                background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
                '&:hover': {
                  background: 'linear-gradient(135deg, #5a6fd8 0%, #6a4190 100%)',
                },
              }}
            >
              {loading ? (
                <CircularProgress size={24} color="inherit" />
              ) : (
                'Sign In'
              )}
            </Button>
          </form>

          <Divider sx={{ my: 3 }}>OR</Divider>

          {/* Demo Accounts */}
          <Box sx={{ mt: 3 }}>
            <Typography
              variant="body2"
              color="text.secondary"
              sx={{ mb: 2, textAlign: 'center' }}
            >
              Use demo account:
            </Typography>
            <Box sx={{ display: 'flex', gap: 1, flexDirection: 'column' }}>
              <Button
                variant="outlined"
                size="small"
                onClick={() => {
                  // Auto-fill admin credentials
                  const event = new Event('input', { bubbles: true });
                  const emailField = document.querySelector('input[name="email"]') as HTMLInputElement;
                  const passwordField = document.querySelector('input[name="password"]') as HTMLInputElement;
                  if (emailField && passwordField) {
                    emailField.value = 'admin@crm.local';
                    passwordField.value = 'admin123';
                    emailField.dispatchEvent(event);
                    passwordField.dispatchEvent(event);
                  }
                }}
              >
                Admin Account
              </Button>
              <Button
                variant="outlined"
                size="small"
                onClick={() => {
                  // Auto-fill team leader credentials
                  const event = new Event('input', { bubbles: true });
                  const emailField = document.querySelector('input[name="email"]') as HTMLInputElement;
                  const passwordField = document.querySelector('input[name="password"]') as HTMLInputElement;
                  if (emailField && passwordField) {
                    emailField.value = 'leader@crm.local';
                    passwordField.value = 'leader123';
                    emailField.dispatchEvent(event);
                    passwordField.dispatchEvent(event);
                  }
                }}
              >
                Team Leader Account
              </Button>
              <Button
                variant="outlined"
                size="small"
                onClick={() => {
                  // Auto-fill user credentials
                  const event = new Event('input', { bubbles: true });
                  const emailField = document.querySelector('input[name="email"]') as HTMLInputElement;
                  const passwordField = document.querySelector('input[name="password"]') as HTMLInputElement;
                  if (emailField && passwordField) {
                    emailField.value = 'user@crm.local';
                    passwordField.value = 'user123';
                    emailField.dispatchEvent(event);
                    passwordField.dispatchEvent(event);
                  }
                }}
              >
                Regular User Account
              </Button>
            </Box>
          </Box>

          {/* Sign Up Link */}
          <Box sx={{ mt: 3, textAlign: 'center' }}>
            <Typography variant="body2" color="text.secondary">
              Don't have an account?{' '}
              <Link href="/register" underline="hover">
                Sign up
              </Link>
            </Typography>
          </Box>
        </Card>
      </motion.div>
    </Box>
  );
};

export default Login;