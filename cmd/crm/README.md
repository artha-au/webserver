# CRM

- Webserver
- Auth/authz

## Admin

/api/v1/admin/teams (CRUDL)
/api/v1/admin/teams/{teamID}/members (CRUDL)

## Team Leaders

/api/v1/teams (List teams user is leader of)
/api/v1/teams/{teamID}/members (List members - scoped to team)
/api/v1/teams/{teamID}/members/{memberID}/timesheets (List, approve, reject - scoped to team)
/api/v1/teams/{teamID}/rosters (CRUDL - scoped to team)

## Regular users

/api/v1/teams
/api/v1/teams/{teamID}/members - scoped to team
/api/v1/teams/{teamID}/rosters (List/Read - scoped to team)
/api/v1/teams/{teamID}/timesheets (CRUDL - scoped to requesting user)