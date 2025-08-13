import React, { useState } from 'react';
import {
  Box,
  Card,
  CardContent,
  Typography,
  Button,
  Grid,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
  IconButton,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Fab,
  Menu,
  ListItemIcon,
  ListItemText,
  Avatar,
  LinearProgress,
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Send as SubmitIcon,
  Check as ApproveIcon,
  Close as RejectIcon,
  MoreVert as MoreIcon,
  AccessTime as ClockIcon,
  TrendingUp as TrendingUpIcon,
  Schedule as ScheduleIcon,
  CheckCircle as CheckCircleIcon,
  Cancel as CancelIcon,
} from '@mui/icons-material';
import { motion } from 'framer-motion';
import { useAuth } from '../../contexts/AuthContext';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { api } from '../../services/api';
import { useForm } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import toast from 'react-hot-toast';

const timesheetSchema = yup.object({
  date: yup.string().required('Date is required'),
  startTime: yup.string().required('Start time is required'),
  endTime: yup.string().required('End time is required'),
  description: yup.string().required('Description is required'),
  teamId: yup.string().required('Team is required'),
});

interface TimesheetFormData {
  date: string;
  startTime: string;
  endTime: string;
  description: string;
  teamId: string;
}

const Timesheets: React.FC = () => {
  const { user, isTeamLeader } = useAuth();
  const queryClient = useQueryClient();
  
  const [openDialog, setOpenDialog] = useState(false);
  const [editingTimesheet, setEditingTimesheet] = useState<any>(null);
  const [selectedTeam, setSelectedTeam] = useState('all');
  const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
  const [selectedTimesheet, setSelectedTimesheet] = useState<any>(null);

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
  } = useForm<TimesheetFormData>({
    resolver: yupResolver(timesheetSchema),
  });

  // Mock data for demonstration
  const mockTimesheets = [
    {
      id: '1',
      date: '2025-01-13',
      startTime: '09:00',
      endTime: '17:00',
      hours: 8,
      description: 'Frontend development - User dashboard implementation',
      status: 'submitted',
      teamName: 'Development Team',
      teamId: 'team1',
      submittedAt: '2025-01-13T17:05:00Z',
    },
    {
      id: '2',
      date: '2025-01-12',
      startTime: '09:00',
      endTime: '16:30',
      hours: 7.5,
      description: 'Code review and bug fixes',
      status: 'approved',
      teamName: 'Development Team',
      teamId: 'team1',
      approvedAt: '2025-01-12T18:30:00Z',
      approvedBy: 'John Doe',
    },
    {
      id: '3',
      date: '2025-01-11',
      startTime: '10:00',
      endTime: '18:00',
      hours: 8,
      description: 'API integration and testing',
      status: 'rejected',
      teamName: 'Development Team',
      teamId: 'team1',
      rejectedAt: '2025-01-11T19:00:00Z',
      rejectedBy: 'Jane Smith',
      rejectionReason: 'Missing detailed task breakdown',
    },
    {
      id: '4',
      date: '2025-01-10',
      startTime: '09:30',
      endTime: '17:30',
      hours: 8,
      description: 'Database optimization and performance tuning',
      status: 'draft',
      teamName: 'Development Team',
      teamId: 'team1',
    },
  ];

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'draft':
        return 'default';
      case 'submitted':
        return 'warning';
      case 'approved':
        return 'success';
      case 'rejected':
        return 'error';
      default:
        return 'default';
    }
  };

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'draft':
        return <EditIcon fontSize="small" />;
      case 'submitted':
        return <ScheduleIcon fontSize="small" />;
      case 'approved':
        return <CheckCircleIcon fontSize="small" />;
      case 'rejected':
        return <CancelIcon fontSize="small" />;
      default:
        return null;
    }
  };

  const handleMenuClick = (event: React.MouseEvent<HTMLElement>, timesheet: any) => {
    setAnchorEl(event.currentTarget);
    setSelectedTimesheet(timesheet);
  };

  const handleMenuClose = () => {
    setAnchorEl(null);
    setSelectedTimesheet(null);
  };

  const onSubmit = async (data: TimesheetFormData) => {
    try {
      console.log('Creating timesheet:', data);
      toast.success('Timesheet created successfully!');
      setOpenDialog(false);
      reset();
    } catch (error) {
      toast.error('Failed to create timesheet');
    }
  };

  const handleEditTimesheet = (timesheet: any) => {
    setEditingTimesheet(timesheet);
    setOpenDialog(true);
    handleMenuClose();
  };

  const handleSubmitTimesheet = async (timesheetId: string) => {
    try {
      console.log('Submitting timesheet:', timesheetId);
      toast.success('Timesheet submitted for approval!');
      handleMenuClose();
    } catch (error) {
      toast.error('Failed to submit timesheet');
    }
  };

  const handleApproveTimesheet = async (timesheetId: string) => {
    try {
      console.log('Approving timesheet:', timesheetId);
      toast.success('Timesheet approved!');
      handleMenuClose();
    } catch (error) {
      toast.error('Failed to approve timesheet');
    }
  };

  const handleRejectTimesheet = async (timesheetId: string) => {
    try {
      console.log('Rejecting timesheet:', timesheetId);
      toast.success('Timesheet rejected');
      handleMenuClose();
    } catch (error) {
      toast.error('Failed to reject timesheet');
    }
  };

  const totalHours = mockTimesheets
    .filter(t => selectedTeam === 'all' || t.teamId === selectedTeam)
    .reduce((sum, t) => sum + t.hours, 0);

  const approvedHours = mockTimesheets
    .filter(t => t.status === 'approved' && (selectedTeam === 'all' || t.teamId === selectedTeam))
    .reduce((sum, t) => sum + t.hours, 0);

  const pendingHours = mockTimesheets
    .filter(t => t.status === 'submitted' && (selectedTeam === 'all' || t.teamId === selectedTeam))
    .reduce((sum, t) => sum + t.hours, 0);

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
              Timesheets
            </Typography>
            <Typography variant="body1" color="text.secondary">
              Track your work hours and manage time entries
            </Typography>
          </Box>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            onClick={() => setOpenDialog(true)}
            sx={{ borderRadius: 2 }}
          >
            New Timesheet
          </Button>
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
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
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
                    <ClockIcon />
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      Total Hours
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
            transition={{ duration: 0.5, delay: 0.2 }}
          >
            <Card>
              <CardContent>
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
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
                    <CheckCircleIcon />
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      Approved Hours
                    </Typography>
                    <Typography variant="h4" sx={{ fontWeight: 600 }}>
                      {approvedHours}
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
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
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
                    <ScheduleIcon />
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      Pending Hours
                    </Typography>
                    <Typography variant="h4" sx={{ fontWeight: 600 }}>
                      {pendingHours}
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
                <Box sx={{ display: 'flex', alignItems: 'center', mb: 2 }}>
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
                    <TrendingUpIcon />
                  </Box>
                  <Box>
                    <Typography variant="body2" color="text.secondary">
                      This Week
                    </Typography>
                    <Typography variant="h4" sx={{ fontWeight: 600 }}>
                      38.5
                    </Typography>
                  </Box>
                </Box>
              </CardContent>
            </Card>
          </motion.div>
        </Grid>
      </Grid>

      {/* Timesheets Table */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.5 }}
      >
        <Card>
          <Box sx={{ p: 2, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <Typography variant="h6" sx={{ fontWeight: 600 }}>
              Time Entries
            </Typography>
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
          </Box>
          
          <TableContainer>
            <Table>
              <TableHead>
                <TableRow>
                  <TableCell>Date</TableCell>
                  <TableCell>Time</TableCell>
                  <TableCell>Hours</TableCell>
                  <TableCell>Description</TableCell>
                  <TableCell>Team</TableCell>
                  <TableCell>Status</TableCell>
                  <TableCell width={50}></TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {mockTimesheets
                  .filter(timesheet => selectedTeam === 'all' || timesheet.teamId === selectedTeam)
                  .map((timesheet) => (
                    <TableRow key={timesheet.id} hover>
                      <TableCell>
                        <Typography variant="body2" sx={{ fontWeight: 600 }}>
                          {new Date(timesheet.date).toLocaleDateString()}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2">
                          {timesheet.startTime} - {timesheet.endTime}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" sx={{ fontWeight: 600 }}>
                          {timesheet.hours}h
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography 
                          variant="body2" 
                          sx={{ 
                            maxWidth: 300,
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                          }}
                        >
                          {timesheet.description}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Typography variant="body2" color="text.secondary">
                          {timesheet.teamName}
                        </Typography>
                      </TableCell>
                      <TableCell>
                        <Chip
                          icon={getStatusIcon(timesheet.status)}
                          label={timesheet.status.charAt(0).toUpperCase() + timesheet.status.slice(1)}
                          color={getStatusColor(timesheet.status) as any}
                          size="small"
                          sx={{ textTransform: 'capitalize' }}
                        />
                      </TableCell>
                      <TableCell>
                        <IconButton
                          size="small"
                          onClick={(e) => handleMenuClick(e, timesheet)}
                        >
                          <MoreIcon />
                        </IconButton>
                      </TableCell>
                    </TableRow>
                  ))}
              </TableBody>
            </Table>
          </TableContainer>
        </Card>
      </motion.div>

      {/* Action Menu */}
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleMenuClose}
      >
        {selectedTimesheet?.status === 'draft' && (
          <>
            <MenuItem onClick={() => handleEditTimesheet(selectedTimesheet)}>
              <ListItemIcon>
                <EditIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText>Edit</ListItemText>
            </MenuItem>
            <MenuItem onClick={() => handleSubmitTimesheet(selectedTimesheet.id)}>
              <ListItemIcon>
                <SubmitIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText>Submit</ListItemText>
            </MenuItem>
          </>
        )}
        {selectedTimesheet?.status === 'submitted' && isTeamLeader && (
          <>
            <MenuItem onClick={() => handleApproveTimesheet(selectedTimesheet.id)}>
              <ListItemIcon>
                <ApproveIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText>Approve</ListItemText>
            </MenuItem>
            <MenuItem onClick={() => handleRejectTimesheet(selectedTimesheet.id)}>
              <ListItemIcon>
                <RejectIcon fontSize="small" />
              </ListItemIcon>
              <ListItemText>Reject</ListItemText>
            </MenuItem>
          </>
        )}
        <MenuItem onClick={() => console.log('View details', selectedTimesheet?.id)}>
          <ListItemIcon>
            <EditIcon fontSize="small" />
          </ListItemIcon>
          <ListItemText>View Details</ListItemText>
        </MenuItem>
      </Menu>

      {/* Create/Edit Timesheet Dialog */}
      <Dialog open={openDialog} onClose={() => setOpenDialog(false)} maxWidth="sm" fullWidth>
        <DialogTitle>
          {editingTimesheet ? 'Edit Timesheet' : 'New Timesheet'}
        </DialogTitle>
        <form onSubmit={handleSubmit(onSubmit)}>
          <DialogContent>
            <Grid container spacing={2}>
              <Grid item xs={12}>
                <TextField
                  {...register('date')}
                  fullWidth
                  label="Date"
                  type="date"
                  InputLabelProps={{ shrink: true }}
                  error={!!errors.date}
                  helperText={errors.date?.message}
                />
              </Grid>
              <Grid item xs={6}>
                <TextField
                  {...register('startTime')}
                  fullWidth
                  label="Start Time"
                  type="time"
                  InputLabelProps={{ shrink: true }}
                  error={!!errors.startTime}
                  helperText={errors.startTime?.message}
                />
              </Grid>
              <Grid item xs={6}>
                <TextField
                  {...register('endTime')}
                  fullWidth
                  label="End Time"
                  type="time"
                  InputLabelProps={{ shrink: true }}
                  error={!!errors.endTime}
                  helperText={errors.endTime?.message}
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
              <Grid item xs={12}>
                <TextField
                  {...register('description')}
                  fullWidth
                  label="Description"
                  multiline
                  rows={3}
                  placeholder="Describe the work you did..."
                  error={!!errors.description}
                  helperText={errors.description?.message}
                />
              </Grid>
            </Grid>
          </DialogContent>
          <DialogActions sx={{ p: 2 }}>
            <Button onClick={() => setOpenDialog(false)}>
              Cancel
            </Button>
            <Button type="submit" variant="contained">
              {editingTimesheet ? 'Update' : 'Create'}
            </Button>
          </DialogActions>
        </form>
      </Dialog>
    </Box>
  );
};

export default Timesheets;