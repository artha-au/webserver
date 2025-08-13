import React, { useState } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  Grid,
  Avatar,
  AvatarGroup,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  IconButton,
  Menu,
  ListItemIcon,
  ListItemText,
  Divider,
  Paper,
  List,
  ListItem,
  ListItemAvatar,
  ListItemText as MuiListItemText,
  ListItemSecondaryAction,
  Badge,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  MoreVert as MoreIcon,
  Group as GroupIcon,
  Person as PersonIcon,
  AdminPanelSettings as AdminIcon,
  SupervisorAccount as LeaderIcon,
  Schedule as ScheduleIcon,
  Assignment as TaskIcon,
  Email as EmailIcon,
  Phone as PhoneIcon,
} from '@mui/icons-material';
import { motion } from 'framer-motion';
import { useAuth } from '../../contexts/AuthContext';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../services/api';
import { useForm } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import toast from 'react-hot-toast';

const teamSchema = yup.object({
  name: yup.string().required('Team name is required'),
  description: yup.string().required('Description is required'),
});

const memberSchema = yup.object({
  email: yup.string().email('Invalid email').required('Email is required'),
  name: yup.string().required('Name is required'),
  role: yup.string().required('Role is required'),
});

interface TeamFormData {
  name: string;
  description: string;
}

interface MemberFormData {
  email: string;
  name: string;
  role: string;
}

const Teams: React.FC = () => {
  const { user, isAdmin, isTeamLeader } = useAuth();
  
  const [openTeamDialog, setOpenTeamDialog] = useState(false);
  const [openMemberDialog, setOpenMemberDialog] = useState(false);
  const [selectedTeam, setSelectedTeam] = useState<any>(null);
  const [editingTeam, setEditingTeam] = useState<any>(null);
  const [editingMember, setEditingMember] = useState<any>(null);
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [menuType, setMenuType] = useState<'team' | 'member'>('team');
  const [menuTarget, setMenuTarget] = useState<any>(null);

  // Fetch user's teams
  const { data: teams } = useQuery({
    queryKey: ['myTeams'],
    queryFn: async () => {
      const response = await api.teams.listMyTeams();
      return response.data;
    },
  });

  const teamForm = useForm<TeamFormData>({
    resolver: yupResolver(teamSchema),
  });

  const memberForm = useForm<MemberFormData>({
    resolver: yupResolver(memberSchema),
  });

  // Mock team data for demonstration
  const mockTeams = [
    {
      id: 'team1',
      name: 'Development Team',
      description: 'Frontend and backend development team responsible for core product features',
      memberCount: 8,
      role: 'leader',
      members: [
        {
          id: 'user1',
          name: 'John Doe',
          email: 'john@company.com',
          role: 'member',
          avatar: 'JD',
          status: 'active',
          joinedDate: '2024-01-15',
        },
        {
          id: 'user2',
          name: 'Jane Smith',
          email: 'jane@company.com',
          role: 'member',
          avatar: 'JS',
          status: 'active',
          joinedDate: '2024-02-01',
        },
        {
          id: 'user3',
          name: 'Mike Johnson',
          email: 'mike@company.com',
          role: 'member',
          avatar: 'MJ',
          status: 'inactive',
          joinedDate: '2023-12-10',
        },
      ],
    },
    {
      id: 'team2',
      name: 'Support Team',
      description: 'Customer support and technical assistance team',
      memberCount: 5,
      role: 'member',
      members: [
        {
          id: 'user4',
          name: 'Sarah Wilson',
          email: 'sarah@company.com',
          role: 'leader',
          avatar: 'SW',
          status: 'active',
          joinedDate: '2023-11-20',
        },
        {
          id: 'user5',
          name: 'Robert Brown',
          email: 'robert@company.com',
          role: 'member',
          avatar: 'RB',
          status: 'active',
          joinedDate: '2024-03-05',
        },
      ],
    },
  ];

  const handleMenuClick = (event: React.MouseEvent<HTMLElement>, type: 'team' | 'member', target: any) => {
    setAnchorEl(event.currentTarget);
    setMenuType(type);
    setMenuTarget(target);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
    setMenuTarget(null);
  };

  const onTeamSubmit = async (data: TeamFormData) => {
    try {
      console.log('Creating/updating team:', data);
      toast.success(editingTeam ? 'Team updated successfully!' : 'Team created successfully!');
      setOpenTeamDialog(false);
      teamForm.reset();
      setEditingTeam(null);
    } catch (error) {
      toast.error('Failed to save team');
    }
  };

  const onMemberSubmit = async (data: MemberFormData) => {
    try {
      console.log('Adding/updating member:', data);
      toast.success(editingMember ? 'Member updated successfully!' : 'Member added successfully!');
      setOpenMemberDialog(false);
      memberForm.reset();
      setEditingMember(null);
    } catch (error) {
      toast.error('Failed to save member');
    }
  };

  const getRoleColor = (role: string) => {
    switch (role) {
      case 'leader':
        return 'primary';
      case 'admin':
        return 'error';
      default:
        return 'default';
    }
  };

  const getRoleIcon = (role: string) => {
    switch (role) {
      case 'leader':
        return <LeaderIcon fontSize="small" />;
      case 'admin':
        return <AdminIcon fontSize="small" />;
      default:
        return <PersonIcon fontSize="small" />;
    }
  };

  return (
    <Box>
      {/* Header */}
      <motion.div
        initial={{ opacity: 0, y: -20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 3 }}>
          <Box>
            <Typography variant="h4" sx={{ fontWeight: 600, mb: 1 }}>
              My Teams
            </Typography>
            <Typography variant="body1" color="text.secondary">
              Manage your teams and team members
            </Typography>
          </Box>
          {(isAdmin || isTeamLeader) && (
            <Button
              variant="contained"
              startIcon={<AddIcon />}
              onClick={() => setOpenTeamDialog(true)}
              sx={{ borderRadius: 2 }}
            >
              New Team
            </Button>
          )}
        </Box>
      </motion.div>

      {/* Teams Grid */}
      <Grid container spacing={3}>
        {mockTeams.map((team, index) => (
          <Grid item xs={12} md={6} key={team.id}>
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: index * 0.1 }}
            >
              <Card
                sx={{
                  height: '100%',
                  cursor: 'pointer',
                  '&:hover': {
                    transform: 'translateY(-4px)',
                    transition: 'transform 0.3s ease',
                  },
                }}
                onClick={() => setSelectedTeam(team)}
              >
                <CardContent>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 2 }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                      <Box
                        sx={{
                          width: 48,
                          height: 48,
                          borderRadius: 2,
                          backgroundColor: 'primary.main',
                          color: 'white',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          fontWeight: 600,
                        }}
                      >
                        <GroupIcon />
                      </Box>
                      <Box>
                        <Typography variant="h6" sx={{ fontWeight: 600 }}>
                          {team.name}
                        </Typography>
                        <Chip
                          icon={getRoleIcon(team.role)}
                          label={team.role.charAt(0).toUpperCase() + team.role.slice(1)}
                          color={getRoleColor(team.role) as any}
                          size="small"
                        />
                      </Box>
                    </Box>
                    <IconButton
                      size="small"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleMenuClick(e, 'team', team);
                      }}
                    >
                      <MoreIcon />
                    </IconButton>
                  </Box>

                  <Typography variant="body2" color="text.secondary" sx={{ mb: 3 }}>
                    {team.description}
                  </Typography>

                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
                      <AvatarGroup max={4}>
                        {team.members.map((member) => (
                          <Avatar key={member.id} sx={{ width: 32, height: 32, bgcolor: 'primary.main' }}>
                            {member.avatar}
                          </Avatar>
                        ))}
                      </AvatarGroup>
                      <Typography variant="body2" color="text.secondary">
                        {team.memberCount} members
                      </Typography>
                    </Box>
                  </Box>

                  <Box sx={{ display: 'flex', gap: 1, mt: 2 }}>
                    <Chip
                      icon={<ScheduleIcon />}
                      label="5 active rosters"
                      size="small"
                      variant="outlined"
                    />
                    <Chip
                      icon={<TaskIcon />}
                      label="12 pending timesheets"
                      size="small"
                      variant="outlined"
                    />
                  </Box>
                </CardContent>
              </Card>
            </motion.div>
          </Grid>
        ))}
      </Grid>

      {/* Team Detail Dialog */}
      <Dialog open={Boolean(selectedTeam)} onClose={() => setSelectedTeam(null)} maxWidth="md" fullWidth>
        <DialogTitle>
          <Box sx={{ display: 'flex', alignItems: 'center', gap: 2 }}>
            <Avatar sx={{ bgcolor: 'primary.main' }}>
              <GroupIcon />
            </Avatar>
            <Box>
              <Typography variant="h6">{selectedTeam?.name}</Typography>
              <Typography variant="body2" color="text.secondary">
                {selectedTeam?.memberCount} members
              </Typography>
            </Box>
          </Box>
        </DialogTitle>
        <DialogContent>
          <Typography variant="body1" sx={{ mb: 3 }}>
            {selectedTeam?.description}
          </Typography>

          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              Team Members
            </Typography>
            {(isAdmin || selectedTeam?.role === 'leader') && (
              <Button
                variant="outlined"
                size="small"
                startIcon={<AddIcon />}
                onClick={() => setOpenMemberDialog(true)}
              >
                Add Member
              </Button>
            )}
          </Box>

          <List>
            {selectedTeam?.members.map((member: any) => (
              <ListItem key={member.id}>
                <ListItemAvatar>
                  <Avatar sx={{ bgcolor: 'primary.main' }}>
                    {member.avatar}
                  </Avatar>
                </ListItemAvatar>
                <MuiListItemText
                  primary={
                    <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                      <Typography variant="body1" sx={{ fontWeight: 600 }}>
                        {member.name}
                      </Typography>
                      <Chip
                        icon={getRoleIcon(member.role)}
                        label={member.role.charAt(0).toUpperCase() + member.role.slice(1)}
                        color={getRoleColor(member.role) as any}
                        size="small"
                      />
                      {member.status === 'inactive' && (
                        <Chip label="Inactive" size="small" variant="outlined" />
                      )}
                    </Box>
                  }
                  secondary={
                    <Box>
                      <Box sx={{ display: 'flex', alignItems: 'center', gap: 0.5, mb: 0.5 }}>
                        <EmailIcon fontSize="small" />
                        <Typography variant="body2">{member.email}</Typography>
                      </Box>
                      <Typography variant="caption" color="text.secondary">
                        Joined {new Date(member.joinedDate).toLocaleDateString()}
                      </Typography>
                    </Box>
                  }
                />
                <ListItemSecondaryAction>
                  <IconButton
                    edge="end"
                    onClick={(e) => handleMenuClick(e, 'member', member)}
                  >
                    <MoreIcon />
                  </IconButton>
                </ListItemSecondaryAction>
              </ListItem>
            ))}
          </List>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSelectedTeam(null)}>Close</Button>
        </DialogActions>
      </Dialog>

      {/* Action Menu */}
      <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={handleMenuClose}>
        {menuType === 'team' && (
          <>
            <MenuItem onClick={() => { setEditingTeam(menuTarget); setOpenTeamDialog(true); handleMenuClose(); }}>
              <ListItemIcon>
                <EditIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText>Edit Team</ListItemText>
            </MenuItem>
            <MenuItem onClick={() => { console.log('View analytics', menuTarget?.id); handleMenuClose(); }}>
              <ListItemIcon>
                <TaskIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText>View Analytics</ListItemText>
            </MenuItem>
            {isAdmin && (
              <MenuItem onClick={() => { console.log('Delete team', menuTarget?.id); handleMenuClose(); }}>
                <ListItemIcon>
                  <DeleteIcon fontSize="small" />
                </ListItemIcon>
                <ListItemText>Delete Team</ListItemText>
              </MenuItem>
            )}
          </>
        )}
        {menuType === 'member' && (
          <>
            <MenuItem onClick={() => { setEditingMember(menuTarget); setOpenMemberDialog(true); handleMenuClose(); }}>
              <ListItemIcon>
                <EditIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText>Edit Member</ListItemText>
            </MenuItem>
            <MenuItem onClick={() => { console.log('View member timesheets', menuTarget?.id); handleMenuClose(); }}>
              <ListItemIcon>
                <ScheduleIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText>View Timesheets</ListItemText>
            </MenuItem>
            {(isAdmin || selectedTeam?.role === 'leader') && (
              <MenuItem onClick={() => { console.log('Remove member', menuTarget?.id); handleMenuClose(); }}>
                <ListItemIcon>
                  <DeleteIcon fontSize="small" />
                </ListItemIcon>
                <ListItemText>Remove Member</ListItemText>
              </MenuItem>
            )}
          </>
        )}
      </Menu>

      {/* Create/Edit Team Dialog */}
      <Dialog open={openTeamDialog} onClose={() => setOpenTeamDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          {editingTeam ? 'Edit Team' : 'New Team'}
        </DialogTitle>
        <form onSubmit={teamForm.handleSubmit(onTeamSubmit)}>
          <DialogContent>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  {...teamForm.register('name')}
                  fullWidth
                  label="Team Name"
                  placeholder="Enter team name..."
                  error={!!teamForm.formState.errors.name}
                  helperText={teamForm.formState.errors.name?.message}
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  {...teamForm.register('description')}
                  fullWidth
                  label="Description"
                  multiline
                  rows={3}
                  placeholder="Describe the team's purpose and responsibilities..."
                  error={!!teamForm.formState.errors.description}
                  helperText={teamForm.formState.errors.description?.message}
                />
              </Grid>
            </Grid>
          </DialogContent>
          <DialogActions sx={{ p: 2 }}>
            <Button onClick={() => setOpenTeamDialog(false)}>Cancel</Button>
            <Button type="submit" variant="contained">
              {editingTeam ? 'Update' : 'Create'}
            </Button>
          </DialogActions>
        </form>
      </Dialog>

      {/* Add/Edit Member Dialog */}
      <Dialog open={openMemberDialog} onClose={() => setOpenMemberDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          {editingMember ? 'Edit Member' : 'Add Member'}
        </DialogTitle>
        <form onSubmit={memberForm.handleSubmit(onMemberSubmit)}>
          <DialogContent>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  {...memberForm.register('name')}
                  fullWidth
                  label="Full Name"
                  placeholder="Enter member's full name..."
                  error={!!memberForm.formState.errors.name}
                  helperText={memberForm.formState.errors.name?.message}
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  {...memberForm.register('email')}
                  fullWidth
                  label="Email Address"
                  placeholder="Enter member's email..."
                  error={!!memberForm.formState.errors.email}
                  helperText={memberForm.formState.errors.email?.message}
                />
              </Grid>
              <Grid item xs={12}>
                <FormControl fullWidth error={!!memberForm.formState.errors.role}>
                  <InputLabel>Role</InputLabel>
                  <Select {...memberForm.register('role')} label="Role">
                    <MenuItem value="member">Member</MenuItem>
                    <MenuItem value="leader">Team Leader</MenuItem>
                    {isAdmin && <MenuItem value="admin">Admin</MenuItem>}
                  </Select>
                </FormControl>
              </Grid>
            </Grid>
          </DialogContent>
          <DialogActions sx={{ p: 2 }}>
            <Button onClick={() => setOpenMemberDialog(false)}>Cancel</Button>
            <Button type="submit" variant="contained">
              {editingMember ? 'Update' : 'Add'}
            </Button>
          </DialogActions>
        </form>
      </Dialog>
    </Box>
  );
};

export default Teams;