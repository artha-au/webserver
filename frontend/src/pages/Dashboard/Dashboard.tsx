import React, { useState } from 'react';
import {
  Box,
  Grid,
  Card,
  CardContent,
  Typography,
  Button,
  Avatar,
  AvatarGroup,
  Chip,
  LinearProgress,
  IconButton,
  Paper,
  List,
  ListItem,
  ListItemAvatar,
  ListItemText,
  ListItemSecondaryAction,
} from '@mui/material';
import {
  TrendingUp as TrendingUpIcon,
  TrendingDown as TrendingDownIcon,
  AccessTime as ClockIcon,
  CheckCircle as CheckIcon,
  Schedule as ScheduleIcon,
  Group as TeamIcon,
  MoreVert as MoreIcon,
  ArrowForward as ArrowIcon,
} from '@mui/icons-material';
import { motion } from 'framer-motion';
import { useAuth } from '../../contexts/AuthContext';
import { api } from '../../services/api';
import { useQuery } from '@tanstack/react-query';

interface StatCard {
  title: string;
  value: string | number;
  change: number;
  icon: React.ReactNode;
  color: string;
}

const Dashboard: React.FC = () => {
  const { user, isAdmin, isTeamLeader } = useAuth();
  const [selectedPeriod, setSelectedPeriod] = useState('week');

  // Fetch user's teams
  const { data: teams } = useQuery({
    queryKey: ['myTeams'],
    queryFn: async () => {
      const response = await api.teams.listMyTeams();
      return response.data;
    },
  });

  // Suppress unused variable warning
  void teams;

  // Mock data for demonstration
  const statCards: StatCard[] = [
    {
      title: 'Total Hours This Week',
      value: '38.5',
      change: 12,
      icon: <ClockIcon />,
      color: '#0073ea',
    },
    {
      title: 'Pending Timesheets',
      value: 3,
      change: -25,
      icon: <ScheduleIcon />,
      color: '#fdab3d',
    },
    {
      title: 'Approved Hours',
      value: '152',
      change: 8,
      icon: <CheckIcon />,
      color: '#00ca72',
    },
    {
      title: 'Team Members',
      value: 12,
      change: 0,
      icon: <TeamIcon />,
      color: '#579bfc',
    },
  ];

  const recentActivities = [
    {
      id: 1,
      user: 'John Doe',
      action: 'submitted timesheet',
      time: '2 hours ago',
      status: 'pending',
    },
    {
      id: 2,
      user: 'Jane Smith',
      action: 'approved roster',
      time: '4 hours ago',
      status: 'success',
    },
    {
      id: 3,
      user: 'Mike Johnson',
      action: 'created new team',
      time: '1 day ago',
      status: 'info',
    },
    {
      id: 4,
      user: 'Sarah Wilson',
      action: 'rejected timesheet',
      time: '2 days ago',
      status: 'error',
    },
  ];

  const upcomingShifts = [
    {
      id: 1,
      date: 'Today',
      time: '09:00 - 17:00',
      team: 'Development Team',
      members: ['JD', 'AS', 'MK'],
    },
    {
      id: 2,
      date: 'Tomorrow',
      time: '10:00 - 18:00',
      team: 'Support Team',
      members: ['SW', 'RB', 'LT'],
    },
    {
      id: 3,
      date: 'Wed, Jan 17',
      time: '09:00 - 17:00',
      team: 'Development Team',
      members: ['JD', 'MK'],
    },
  ];

  return (
    <Box>
      {/* Welcome Section */}
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <Box sx={{ mb: 4 }}>
          <Typography variant="h4" sx={{ fontWeight: 600, mb: 1 }}>
            Welcome back, {user?.name}! 👋
          </Typography>
          <Typography variant="body1" color="text.secondary">
            Here's what's happening with your teams today.
          </Typography>
        </Box>
      </motion.div>

      {/* Stats Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        {statCards.map((stat, index) => (
          <Grid item xs={12} sm={6} md={3} key={index}>
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
            >
              <Card
                sx={{
                  height: '100%',
                  position: 'relative',
                  overflow: 'visible',
                  '&:hover': {
                    transform: 'translateY(-4px)',
                    transition: 'transform 0.3s ease',
                  },
                }}
              >
                <CardContent>
                  <Box sx={{ display: 'flex', alignItems: 'flex-start', mb: 2 }}>
                    <Box
                      sx={{
                        width: 48,
                        height: 48,
                        borderRadius: 2,
                        backgroundColor: `${stat.color}20`,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: stat.color,
                        mr: 2,
                      }}
                    >
                      {stat.icon}
                    </Box>
                    <Box sx={{ flexGrow: 1 }}>
                      <Typography
                        variant="body2"
                        color="text.secondary"
                        sx={{ mb: 0.5 }}
                      >
                        {stat.title}
                      </Typography>
                      <Typography variant="h4" sx={{ fontWeight: 600 }}>
                        {stat.value}
                      </Typography>
                    </Box>
                  </Box>
                  <Box sx={{ display: 'flex', alignItems: 'center' }}>
                    {stat.change > 0 ? (
                      <TrendingUpIcon
                        sx={{ color: 'success.main', fontSize: 20, mr: 0.5 }}
                      />
                    ) : stat.change < 0 ? (
                      <TrendingDownIcon
                        sx={{ color: 'error.main', fontSize: 20, mr: 0.5 }}
                      />
                    ) : null}
                    <Typography
                      variant="body2"
                      sx={{
                        color:
                          stat.change > 0
                            ? 'success.main'
                            : stat.change < 0
                            ? 'error.main'
                            : 'text.secondary',
                        fontWeight: 500,
                      }}
                    >
                      {Math.abs(stat.change)}% from last {selectedPeriod}
                    </Typography>
                  </Box>
                </CardContent>
              </Card>
            </motion.div>
          </Grid>
        ))}
      </Grid>

      <Grid container spacing={3}>
        {/* Recent Activity */}
        <Grid item xs={12} md={8}>
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.4 }}
          >
            <Card>
              <Box
                sx={{
                  p: 2,
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  borderBottom: '1px solid',
                  borderColor: 'divider',
                }}
              >
                <Typography variant="h6" sx={{ fontWeight: 600 }}>
                  Recent Activity
                </Typography>
                <Button
                  size="small"
                  endIcon={<ArrowIcon />}
                  sx={{ textTransform: 'none' }}
                >
                  View All
                </Button>
              </Box>
              <List sx={{ p: 0 }}>
                {recentActivities.map((activity, index) => (
                  <ListItem
                    key={activity.id}
                    sx={{
                      borderBottom:
                        index < recentActivities.length - 1
                          ? '1px solid'
                          : 'none',
                      borderColor: 'divider',
                      '&:hover': {
                        backgroundColor: 'action.hover',
                      },
                    }}
                  >
                    <ListItemAvatar>
                      <Avatar sx={{ bgcolor: 'primary.main' }}>
                        {activity.user
                          .split(' ')
                          .map((n) => n[0])
                          .join('')}
                      </Avatar>
                    </ListItemAvatar>
                    <ListItemText
                      primary={
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <Typography variant="body2" sx={{ fontWeight: 600 }}>
                            {activity.user}
                          </Typography>
                          <Typography variant="body2" color="text.secondary">
                            {activity.action}
                          </Typography>
                        </Box>
                      }
                      secondary={activity.time}
                    />
                    <ListItemSecondaryAction>
                      <Chip
                        label={activity.status}
                        size="small"
                        color={
                          activity.status === 'success'
                            ? 'success'
                            : activity.status === 'error'
                            ? 'error'
                            : activity.status === 'pending'
                            ? 'warning'
                            : 'info'
                        }
                        sx={{ textTransform: 'capitalize' }}
                      />
                    </ListItemSecondaryAction>
                  </ListItem>
                ))}
              </List>
            </Card>
          </motion.div>
        </Grid>

        {/* Upcoming Shifts */}
        <Grid item xs={12} md={4}>
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            animate={{ opacity: 1, x: 0 }}
            transition={{ duration: 0.5, delay: 0.5 }}
          >
            <Card>
              <Box
                sx={{
                  p: 2,
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  borderBottom: '1px solid',
                  borderColor: 'divider',
                }}
              >
                <Typography variant="h6" sx={{ fontWeight: 600 }}>
                  Upcoming Shifts
                </Typography>
                <IconButton size="small">
                  <MoreIcon />
                </IconButton>
              </Box>
              <Box sx={{ p: 2 }}>
                {upcomingShifts.map((shift, index) => (
                  <Paper
                    key={shift.id}
                    elevation={0}
                    sx={{
                      p: 2,
                      mb: index < upcomingShifts.length - 1 ? 2 : 0,
                      backgroundColor: 'grey.50',
                      borderLeft: '4px solid',
                      borderColor: 'primary.main',
                    }}
                  >
                    <Box sx={{ display: 'flex', justifyContent: 'space-between', mb: 1 }}>
                      <Typography variant="body2" sx={{ fontWeight: 600 }}>
                        {shift.date}
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {shift.time}
                      </Typography>
                    </Box>
                    <Typography variant="body2" color="text.secondary" sx={{ mb: 1 }}>
                      {shift.team}
                    </Typography>
                    <AvatarGroup max={4} sx={{ justifyContent: 'flex-start' }}>
                      {shift.members.map((member, idx) => (
                        <Avatar
                          key={idx}
                          sx={{
                            width: 24,
                            height: 24,
                            fontSize: '0.75rem',
                            bgcolor: 'primary.main',
                          }}
                        >
                          {member}
                        </Avatar>
                      ))}
                    </AvatarGroup>
                  </Paper>
                ))}
              </Box>
            </Card>
          </motion.div>
        </Grid>

        {/* Team Performance (for leaders/admins) */}
        {(isAdmin || isTeamLeader) && (
          <Grid item xs={12}>
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.6 }}
            >
              <Card>
                <Box
                  sx={{
                    p: 2,
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    borderBottom: '1px solid',
                    borderColor: 'divider',
                  }}
                >
                  <Typography variant="h6" sx={{ fontWeight: 600 }}>
                    Team Performance
                  </Typography>
                  <Box sx={{ display: 'flex', gap: 1 }}>
                    <Chip
                      label="Week"
                      onClick={() => setSelectedPeriod('week')}
                      color={selectedPeriod === 'week' ? 'primary' : 'default'}
                      variant={selectedPeriod === 'week' ? 'filled' : 'outlined'}
                    />
                    <Chip
                      label="Month"
                      onClick={() => setSelectedPeriod('month')}
                      color={selectedPeriod === 'month' ? 'primary' : 'default'}
                      variant={selectedPeriod === 'month' ? 'filled' : 'outlined'}
                    />
                    <Chip
                      label="Quarter"
                      onClick={() => setSelectedPeriod('quarter')}
                      color={selectedPeriod === 'quarter' ? 'primary' : 'default'}
                      variant={selectedPeriod === 'quarter' ? 'filled' : 'outlined'}
                    />
                  </Box>
                </Box>
                <Box sx={{ p: 3 }}>
                  <Grid container spacing={3}>
                    {['Development Team', 'Support Team', 'Sales Team'].map((team) => (
                      <Grid item xs={12} sm={4} key={team}>
                        <Box sx={{ mb: 2 }}>
                          <Typography variant="body2" sx={{ fontWeight: 600, mb: 1 }}>
                            {team}
                          </Typography>
                          <Box sx={{ display: 'flex', alignItems: 'center', mb: 1 }}>
                            <Typography variant="h5" sx={{ fontWeight: 600, mr: 1 }}>
                              87%
                            </Typography>
                            <Typography variant="body2" color="success.main">
                              +5%
                            </Typography>
                          </Box>
                          <LinearProgress
                            variant="determinate"
                            value={87}
                            sx={{
                              height: 8,
                              borderRadius: 4,
                              backgroundColor: 'grey.200',
                              '& .MuiLinearProgress-bar': {
                                borderRadius: 4,
                                backgroundColor: 'success.main',
                              },
                            }}
                          />
                        </Box>
                      </Grid>
                    ))}
                  </Grid>
                </Box>
              </Card>
            </motion.div>
          </Grid>
        )}
      </Grid>
    </Box>
  );
};

export default Dashboard;