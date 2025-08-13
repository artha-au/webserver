# CRM Package

A comprehensive Customer Relationship Management (CRM) package for managing teams, rosters, and timesheets in a PostgreSQL-backed Go application.

## Features

### Teams Management
- **Team Creation**: Create teams with designated leaders
- **Team Members**: Add/remove team members with roles (leader/member)
- **Team Statistics**: Get insights into team performance and activity

### Roster Management
- **Schedule Creation**: Team leaders can create work rosters
- **Shift Management**: Define individual shifts with start/end times
- **Roster Publishing**: Control roster visibility (draft/published/archived)
- **Overlap Prevention**: Automatic validation to prevent conflicting rosters

### Timesheet Management
- **Time Tracking**: Team members can create and submit timesheets
- **Review Process**: Team leaders can approve/reject submitted timesheets
- **Status Tracking**: Full audit trail of timesheet lifecycle
- **Statistics**: Individual and team-level time tracking metrics

## Architecture

### Core Components

```
pkg/crm/
├── models.go          # Data models and request/response types
├── store.go           # Main store interface and team operations
├── store_members.go   # Team member operations
├── store_rosters.go   # Roster and shift operations
├── store_timesheets.go # Timesheet operations
├── migrations.go      # Database migrations
├── crm.go            # Main CRM service with business logic
└── README.md         # This file
```

### Database Schema

#### Teams (`crm_teams`)
- Team information and metadata
- Leader assignment
- Active/inactive status

#### Team Members (`crm_team_members`)
- User-team relationships
- Role assignments (leader/member)
- Join dates and status

#### Rosters (`crm_rosters`)
- Work schedules per team
- Date ranges and descriptions
- Status management (draft/published/archived)

#### Roster Shifts (`crm_roster_shifts`)
- Individual work shifts
- User assignments
- Time periods and roles

#### Timesheets (`crm_timesheets`)
- Daily time entries
- Review workflow
- Hours tracking and descriptions

## Usage

### Basic Setup

```go
package main

import (
    "database/sql"
    "log"
    
    "github.com/artha-au/webserver/pkg/crm"
    _ "github.com/lib/pq"
)

func main() {
    // Open database connection
    db, err := sql.Open("postgres", "your-connection-string")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create CRM service with auto-migration
    config := crm.DefaultConfig()
    crmService, err := crm.New(db, config, nil)
    if err != nil {
        log.Fatal("Failed to initialize CRM:", err)
    }
    defer crmService.Close()
    
    // Use the service
    store := crmService.Store()
    // ... perform operations
}
```

### Team Operations

```go
// Create a team
teamReq := &crm.CreateTeamRequest{
    Name:        "Development Team",
    Description: "Software development team",
    LeaderID:    "user-123",
}
team, err := crmService.CreateTeamWithValidation(ctx, teamReq, "admin-user")

// Add team members
memberReq := &crm.AddTeamMemberRequest{
    UserID: "user-456",
    Role:   crm.RoleMember,
}
member, err := store.AddTeamMember(ctx, team.ID, memberReq, "admin-user")

// List user's teams
teams, err := store.ListUserTeams(ctx, "user-123")
```

### Roster Operations

```go
// Create a roster (team leader only)
rosterReq := &crm.CreateRosterRequest{
    Name:        "Weekly Schedule",
    Description: "Week of March 1-7",
    StartDate:   time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
    EndDate:     time.Date(2024, 3, 7, 23, 59, 59, 0, time.UTC),
}
roster, err := crmService.CreateRosterWithValidation(ctx, teamID, rosterReq, "leader-user")

// Add shifts to roster
shiftReq := &crm.CreateRosterShiftRequest{
    UserID:    "user-456",
    StartTime: time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC),
    EndTime:   time.Date(2024, 3, 1, 17, 0, 0, 0, time.UTC),
    Role:      "Developer",
    Notes:     "Morning standup at 9:15",
}
shift, err := store.CreateRosterShift(ctx, roster.ID, shiftReq)

// Publish roster
err = store.PublishRoster(ctx, roster.ID)
```

### Timesheet Operations

```go
// Create timesheet
timesheetReq := &crm.CreateTimesheetRequest{
    Date:        time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC),
    Hours:       8.0,
    Description: "Worked on user authentication feature",
}
timesheet, err := crmService.CreateTimesheetWithValidation(ctx, userID, teamID, timesheetReq)

// Submit for review
err = crmService.SubmitTimesheetWithValidation(ctx, userID, timesheet.ID)

// Review timesheet (team leader only)
reviewReq := &crm.ReviewTimesheetRequest{
    Status:      crm.TimesheetStatusApproved,
    ReviewNotes: "Good work on the authentication feature",
}
err = crmService.ReviewTimesheetWithValidation(ctx, leaderID, timesheet.ID, reviewReq)
```

## Business Rules

### Team Management
- Each team must have exactly one leader
- Users can be members of multiple teams
- Team names must be unique
- Only team leaders can manage team rosters

### Roster Management
- Only team leaders can create/modify rosters
- Rosters cannot have overlapping date ranges within the same team
- Published rosters cannot be modified (must archive and create new)
- Shifts cannot overlap for the same user

### Timesheet Management
- Users can only create timesheets for teams they belong to
- One timesheet per user per date per team
- Only draft timesheets can be edited
- Only team leaders can review submitted timesheets
- Approved/rejected timesheets cannot be modified

## Validation

### Automatic Validations
- **Date Ranges**: End dates must be after start dates
- **Hours**: Must be between 0-24 hours per day
- **Text Limits**: Names (255 chars), descriptions (1000 chars)
- **Status Transitions**: Proper workflow enforcement
- **Access Control**: Role-based operation permissions

### Custom Business Logic
- **Overlap Prevention**: Rosters and shifts cannot overlap
- **Team Membership**: Operations restricted to team members
- **Leadership Requirements**: Certain operations require team leader role
- **Historical Limits**: Timesheets limited to reasonable date ranges

## Database Migrations

The package includes automatic database migrations:

```go
// Manual migration control
migrator := crm.NewMigrator(db, logger)

// Initialize schema
err := migrator.Init(ctx)

// Reset all CRM data (use with caution!)
err := migrator.Reset(ctx)
```

### Migration History
1. **v1**: Create teams table
2. **v2**: Create team members table
3. **v3**: Create rosters table
4. **v4**: Create roster shifts table
5. **v5**: Create timesheets table
6. **v6**: Add performance indexes

## API Integration

The CRM package is designed to work seamlessly with HTTP APIs:

```go
// Example API handler
func (api *APIHandlers) createTeam(w http.ResponseWriter, r *http.Request) {
    var req crm.CreateTeamRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }
    
    user := auth.GetUserFromContext(r)
    team, err := api.crm.CreateTeamWithValidation(r.Context(), &req, user.ID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    json.NewEncoder(w).Encode(team)
}
```

## Error Handling

The package provides detailed error messages for common scenarios:

- **Access Control**: "user is not a leader of this team"
- **Data Validation**: "hours must be between 0 and 24"
- **Business Rules**: "roster dates overlap with existing roster"
- **Status Conflicts**: "can only review submitted timesheets"

## Performance Considerations

### Indexes
- All foreign keys are indexed
- Common query patterns are optimized
- Date range queries use composite indexes

### Query Patterns
- List operations include pagination
- Statistics queries are optimized
- Batch operations minimize database round-trips

### Recommended Practices
- Use context for request cancellation
- Implement connection pooling for high concurrency
- Consider read replicas for reporting queries
- Monitor slow queries and optimize as needed

## Testing

```go
// Example test setup
func TestCreateTeam(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()
    
    crmService, err := crm.New(db, crm.DefaultConfig(), nil)
    require.NoError(t, err)
    defer crmService.Close()
    
    req := &crm.CreateTeamRequest{
        Name:     "Test Team",
        LeaderID: "user-123",
    }
    
    team, err := crmService.CreateTeamWithValidation(context.Background(), req, "admin")
    require.NoError(t, err)
    assert.Equal(t, "Test Team", team.Name)
}
```

## Security Considerations

### Access Control
- All operations enforce team membership rules
- Role-based permissions for sensitive operations
- User context validation for data access

### Data Validation
- SQL injection prevention through parameterized queries
- Input sanitization and length limits
- Business rule enforcement at service layer

### Audit Trail
- All operations include user attribution
- Timestamp tracking for all changes
- Immutable review history for timesheets