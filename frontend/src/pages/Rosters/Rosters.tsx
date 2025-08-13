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
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Divider,
  Checkbox,
  FormControlLabel,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  MoreVert as MoreIcon,
  CalendarMonth as CalendarIcon,
  AccessTime as TimeIcon,
  Group as GroupIcon,
  Person as PersonIcon,
  Event as EventIcon,
  Schedule as ScheduleIcon,
  Visibility as ViewIcon,
  Today as TodayIcon,
} from '@mui/icons-material';
import { motion } from 'framer-motion';
import { useAuth } from '../../contexts/AuthContext';
import { useQuery } from '@tanstack/react-query';
import { api } from '../../services/api';
import { useForm } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import toast from 'react-hot-toast';

const rosterSchema = yup.object({
  name: yup.string().required('Roster name is required'),
  teamId: yup.string().required('Team is required'),
  startDate: yup.string().required('Start date is required'),
  endDate: yup.string().required('End date is required'),
  description: yup.string(),
});

interface RosterFormData {
  name: string;
  teamId: string;
  startDate: string;
  endDate: string;
  description: string;
}

const Rosters: React.FC = () => {
  const { user, isTeamLeader } = useAuth();
  
  const [openDialog, setOpenDialog] = useState(false);
  const [selectedRoster, setSelectedRoster] = useState<any>(null);
  const [editingRoster, setEditingRoster] = useState<any>(null);
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [menuTarget, setMenuTarget] = useState<any>(null);
  const [selectedTeam, setSelectedTeam] = useState('all');
  const [viewMode, setViewMode] = useState<'grid' | 'table'>('grid');

  // Fetch user's teams
  const { data: teams } = useQuery({
    queryKey: ['myTeams'],
    queryFn: async () => {
      const response = await api.teams.listMyTeams();
      return response.data;
    },
  });

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<RosterFormData>({
    resolver: yupResolver(rosterSchema),
  });

  // Mock roster data
  const mockRosters = [
    {
      id: '1',
      name: 'Development Sprint 24.1',
      team: {
        id: 'team1',
        name: 'Development Team',
      },
      startDate: '2025-01-13',
      endDate: '2025-01-19',
      description: 'Weekly development roster for sprint planning and execution',
      status: 'active',
      shifts: [
        {
          id: 'shift1',
          date: '2025-01-13',
          startTime: '09:00',
          endTime: '17:00',
          assignedMembers: [
            { id: 'user1', name: 'John Doe', avatar: 'JD' },
            { id: 'user2', name: 'Jane Smith', avatar: 'JS' },
          ],
        },
        {
          id: 'shift2',
          date: '2025-01-14',
          startTime: '09:00',
          endTime: '17:00',
          assignedMembers: [
            { id: 'user1', name: 'John Doe', avatar: 'JD' },
            { id: 'user3', name: 'Mike Johnson', avatar: 'MJ' },
          ],
        },
      ],
      totalHours: 64,
      memberCount: 3,
    },
    {
      id: '2',
      name: 'Support Coverage',
      team: {
        id: 'team2',
        name: 'Support Team',
      },
      startDate: '2025-01-06',
      endDate: '2025-01-12',
      description: 'Customer support rotation schedule',
      status: 'completed',
      shifts: [
        {
          id: 'shift3',
          date: '2025-01-06',
          startTime: '08:00',
          endTime: '16:00',
          assignedMembers: [
            { id: 'user4', name: 'Sarah Wilson', avatar: 'SW' },
          ],
        },
        {
          id: 'shift4',
          date: '2025-01-07',
          startTime: '10:00',
          endTime: '18:00',
          assignedMembers: [
            { id: 'user5', name: 'Robert Brown', avatar: 'RB' },
          ],
        },
      ],
      totalHours: 48,
      memberCount: 2,
    },
    {
      id: '3',
      name: 'Holiday Coverage',
      team: {
        id: 'team1',
        name: 'Development Team',
      },
      startDate: '2025-01-20',
      endDate: '2025-01-26',
      description: 'Special roster for holiday period with reduced staffing',
      status: 'draft',
      shifts: [],
      totalHours: 0,
      memberCount: 0,
    },
  ];

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'active':
        return 'success';
      case 'completed':
        return 'default';
      case 'draft':
        return 'warning';
      default:
        return 'default';
    }
  };

  const handleMenuClick = (event: React.MouseEvent<HTMLElement>, roster: any) => {
    setAnchorEl(event.currentTarget);
    setMenuTarget(roster);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
    setMenuTarget(null);
  };

  const onSubmit = async (data: RosterFormData) => {
    try {
      console.log('Creating/updating roster:', data);
      toast.success(editingRoster ? 'Roster updated successfully!' : 'Roster created successfully!');
      setOpenDialog(false);
      reset();
      setEditingRoster(null);
    } catch (error) {
      toast.error('Failed to save roster');
    }
  };

  const filteredRosters = mockRosters.filter(roster => 
    selectedTeam === 'all' || roster.team.id === selectedTeam
  );

  const upcomingRosters = filteredRosters.filter(r => r.status === 'active').length;
  const totalHours = filteredRosters.reduce((sum, r) => sum + r.totalHours, 0);
  const totalMembers = filteredRosters.reduce((sum, r) => sum + r.memberCount, 0);

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
              Rosters
            </Typography>
            <Typography variant="body1" color="text.secondary">
              Manage work schedules and shift assignments
            </Typography>
          </Box>
          <Box sx={{ display: 'flex', gap: 1 }}>
            <FormControl size="small" sx={{ minWidth: 150 }}>
              <InputLabel>Filter by Team</InputLabel>
              <Select
                value={selectedTeam}
                onChange={(e) => setSelectedTeam(e.target.value)}
                label="Filter by Team"
              >
                <MenuItem value="all">All Teams</MenuItem>
                {teams?.map((team: any) => (
                  <MenuItem key={team.id} value={team.id}>
                    {team.name}
                  </MenuItem>
                ))}
              </Select>
            </FormControl>
            {isTeamLeader && (
              <Button
                variant="contained"
                startIcon={<AddIcon />}
                onClick={() => setOpenDialog(true)}
                sx={{ borderRadius: 2 }}
              >
                New Roster
              </Button>
            )}
          </Box>
        </Box>
      </motion.div>

      {/* Stats Cards */}
      <Grid container spacing={3} sx={{ mb: 4 }}>
        <Grid item xs={12} md={3}>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.1 }}
          >
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <Box
                    sx={{
                      width: 48,
                      height: 48,
                      borderRadius: 2,
                      backgroundColor: '#0073ea20',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      color: '#0073ea',
                      mr: 2,
                    }}
                  >
                    <CalendarIcon />
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      Active Rosters
                    </Typography>
                    <Typography variant="h4" sx={{ fontWeight: 600 }}>
                      {upcomingRosters}
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </motion.div>
        </Grid>

        <Grid item xs={12} md={3}>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
          >
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <Box
                    sx={{
                      width: 48,
                      height: 48,
                      borderRadius: 2,
                      backgroundColor: '#00ca7220',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      color: '#00ca72',
                      mr: 2,
                    }}
                  >
                    <TimeIcon />
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      Scheduled Hours
                    </Typography>
                    <Typography variant="h4" sx={{ fontWeight: 600 }}>
                      {totalHours}
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </motion.div>
        </Grid>

        <Grid item xs={12} md={3}>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.3 }}
          >
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <Box
                    sx={{
                      width: 48,
                      height: 48,
                      borderRadius: 2,
                      backgroundColor: '#579bfc20',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      color: '#579bfc',
                      mr: 2,
                    }}
                  >
                    <GroupIcon />
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      Assigned Members
                    </Typography>
                    <Typography variant="h4" sx={{ fontWeight: 600 }}>
                      {totalMembers}
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </motion.div>
        </Grid>

        <Grid item xs={12} md={3}>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.4 }}
          >
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center' }}>
                  <Box
                    sx={{
                      width: 48,
                      height: 48,
                      borderRadius: 2,
                      backgroundColor: '#fdab3d20',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      color: '#fdab3d',
                      mr: 2,
                    }}
                  >
                    <TodayIcon />
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      Today's Shifts
                    </Typography>
                    <Typography variant="h4" sx={{ fontWeight: 600 }}>
                      3
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </motion.div>
        </Grid>
      </Grid>

      {/* Rosters Grid */}
      <Grid container spacing={3}>
        {filteredRosters.map((roster, index) => (
          <Grid item xs={12} md={6} lg={4} key={roster.id}>
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
                onClick={() => setSelectedRoster(roster)}
              >
                <CardContent>
                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', mb: 2 }}>
                    <Box>
                      <Typography variant="h6" sx={{ fontWeight: 600, mb: 1 }}>
                        {roster.name}
                      </Typography>
                      <Chip
                        label={roster.status.charAt(0).toUpperCase() + roster.status.slice(1)}
                        color={getStatusColor(roster.status) as any}
                        size="small"
                        sx={{ mb: 1 }}
                      />
                    </Box>
                    <IconButton
                      size="small"
                      onClick={(e) => {
                        e.stopPropagation();
                        handleMenuClick(e, roster);
                      }}
                    >
                      <MoreIcon />
                    </IconButton>
                  </Box>

                  <Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
                    {roster.team.name}
                  </Typography>

                  <Typography variant="body2" sx={{ mb: 3, minHeight: 40 }}>
                    {roster.description}
                  </Typography>

                  <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, mb: 2 }}>
                    <EventIcon fontSize="small" color="action" />
                    <Typography variant="body2" color="text.secondary">
                      {new Date(roster.startDate).toLocaleDateString()} - {new Date(roster.endDate).toLocaleDateString()}
                    </Typography>
                  </Box>

                  <Divider sx={{ my: 2 }} />

                  <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Box>
                      {roster.shifts.length > 0 ? (
                        <AvatarGroup max={3}>
                          {roster.shifts.flatMap(shift => shift.assignedMembers).slice(0, 3).map((member, idx) => (
                            <Avatar key={`${member.id}-${idx}`} sx={{ width: 28, height: 28, fontSize: '0.75rem', bgcolor: 'primary.main' }}>
                              {member.avatar}
                            </Avatar>
                          ))}
                        </AvatarGroup>
                      ) : (
                        <Typography variant="body2" color="text.secondary">
                          No shifts assigned
                        </Typography>
                      )}
                    </Box>
                    <Box sx={{ textAlign: 'right' }}>
                      <Typography variant="body2" sx={{ fontWeight: 600 }}>
                        {roster.totalHours}h
                      </Typography>
                      <Typography variant="caption" color="text.secondary">
                        {roster.shifts.length} shifts
                      </Typography>
                    </Box>
                  </Box>
                </CardContent>
              </Card>
            </motion.div>
          </Grid>
        ))}
      </Grid>

      {/* Roster Detail Dialog */}
      <Dialog open={Boolean(selectedRoster)} onClose={() => setSelectedRoster(null)} maxWidth="lg" fullWidth>
        <DialogTitle>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Box>
              <Typography variant="h6">{selectedRoster?.name}</Typography>
              <Typography variant="body2" color="text.secondary">
                {selectedRoster?.team.name} • {selectedRoster?.startDate} to {selectedRoster?.endDate}
              </Typography>
            </Box>
            <Chip
              label={selectedRoster?.status?.charAt(0).toUpperCase() + selectedRoster?.status?.slice(1)}
              color={getStatusColor(selectedRoster?.status) as any}
            />
          </Box>
        </DialogTitle>
        <DialogContent>
          <Typography variant="body1" sx={{ mb: 3 }}>
            {selectedRoster?.description}
          </Typography>

          <Typography variant="h6" sx={{ fontWeight: 600, mb: 2 }}>
            Shift Schedule
          </Typography>

          {selectedRoster?.shifts && selectedRoster.shifts.length > 0 ? (
            <TableContainer component={Paper} variant="outlined">
              <Table>
                <TableHead>
                  <TableRow>
                    <TableCell>Date</TableCell>
                    <TableCell>Time</TableCell>
                    <TableCell>Assigned Members</TableCell>
                    <TableCell>Hours</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {selectedRoster.shifts.map((shift: any) => (
                    <TableRow key={shift.id}>
                      <TableCell>
                        <Typography variant="body2" sx={{ fontWeight: 600 }}>
                          {new Date(shift.date).toLocaleDateString()}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2">
                          {shift.startTime} - {shift.endTime}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
                          <AvatarGroup max={3}>
                            {shift.assignedMembers.map((member: any) => (
                              <Avatar key={member.id} sx={{ width: 24, height: 24, fontSize: '0.75rem' }}>
                                {member.avatar}
                              </Avatar>
                            ))}
                          </AvatarGroup>
                          <Typography variant="body2">
                            {shift.assignedMembers.map((m: any) => m.name).join(', ')}
                          </Typography>
                        </Box>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2">
                          {(() => {
                            const start = new Date(`2000-01-01 ${shift.startTime}`);
                            const end = new Date(`2000-01-01 ${shift.endTime}`);
                            const hours = (end.getTime() - start.getTime()) / (1000 * 60 * 60);
                            return `${hours}h`;
                          })()}
                        </Typography>
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          ) : (
            <Paper sx={{ p: 3, textAlign: 'center' }} variant="outlined">
              <ScheduleIcon sx={{ fontSize: 48, color: 'text.secondary', mb: 1 }} />
              <Typography variant="body1" color="text.secondary">
                No shifts scheduled yet
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Add shifts to start managing this roster
              </Typography>
            </Paper>
          )}
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setSelectedRoster(null)}>Close</Button>
          {isTeamLeader && (
            <Button
              variant="contained"
              startIcon={<EditIcon />}
              onClick={() => {
                setEditingRoster(selectedRoster);
                setSelectedRoster(null);
                setOpenDialog(true);
              }}
            >
              Edit Roster
            </Button>
          )}
        </DialogActions>
      </Dialog>

      {/* Action Menu */}
      <Menu anchorEl={anchorEl} open={Boolean(anchorEl)} onClose={handleMenuClose}>
        <MenuItem onClick={() => { setSelectedRoster(menuTarget); handleMenuClose(); }}>
          <ListItemIcon>
            <ViewIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText>View Details</ListItemText>
        </MenuItem>
        {isTeamLeader && (
          <>
            <MenuItem onClick={() => { setEditingRoster(menuTarget); setOpenDialog(true); handleMenuClose(); }}>
              <ListItemIcon>
                <EditIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText>Edit Roster</ListItemText>
            </MenuItem>
            <MenuItem onClick={() => { console.log('Duplicate roster', menuTarget?.id); handleMenuClose(); }}>
              <ListItemIcon>
                <AddIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText>Duplicate</ListItemText>
            </MenuItem>
            <MenuItem onClick={() => { console.log('Delete roster', menuTarget?.id); handleMenuClose(); }}>
              <ListItemIcon>
                <DeleteIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText>Delete</ListItemText>
            </MenuItem>
          </>
        )}
      </Menu>

      {/* Create/Edit Roster Dialog */}
      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          {editingRoster ? 'Edit Roster' : 'New Roster'}
        </DialogTitle>
        <form onSubmit={handleSubmit(onSubmit)}>
          <DialogContent>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  {...register('name')}
                  fullWidth
                  label="Roster Name"
                  placeholder="Enter roster name..."
                  error={!!errors.name}
                  helperText={errors.name?.message}
                />
              </Grid>
              <Grid item xs={12}>
                <FormControl fullWidth error={!!errors.teamId}>
                  <InputLabel>Team</InputLabel>
                  <Select {...register('teamId')} label="Team">
                    {teams?.map((team: any) => (
                      <MenuItem key={team.id} value={team.id}>
                        {team.name}
                      </MenuItem>
                    ))}
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={6}>
                <TextField
                  {...register('startDate')}
                  fullWidth
                  label="Start Date"
                  type="date"
                  InputLabelProps={{ shrink: true }}
                  error={!!errors.startDate}
                  helperText={errors.startDate?.message}
                />
              </Grid>
              <Grid item xs={6}>
                <TextField
                  {...register('endDate')}
                  fullWidth
                  label="End Date"
                  type="date"
                  InputLabelProps={{ shrink: true }}
                  error={!!errors.endDate}
                  helperText={errors.endDate?.message}
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  {...register('description')}
                  fullWidth
                  label="Description"
                  multiline
                  rows={3}
                  placeholder="Describe the roster purpose and schedule..."
                  error={!!errors.description}
                  helperText={errors.description?.message}
                />
              </Grid>
            </Grid>
          </DialogContent>
          <DialogActions sx={{ p: 2 }}>
            <Button onClick={() => setOpenDialog(false)}>Cancel</Button>
            <Button type="submit" variant="contained">
              {editingRoster ? 'Update' : 'Create'}
            </Button>
          </DialogActions>
        </form>
      </Dialog>
    </Box>
  );
};

export default Rosters;