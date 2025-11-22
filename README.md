# terraform-volleygoals

This is the application Module for VolleyGoals, a web application for managing volleyball goals and statistics.
This module is designed to be used with Terraform to provision the necessary infrastructure on AWS.

## Prerequisites

- Terraform 1.0 or later
- AWS CLI configured with appropriate credentials

## Usage


## Database Diagram

The following diagram illustrates the database schema for the Volleyball Goal Management application:

```mermaid
%% Cognito as IdP (no Users table). No Trainings yet. Comments/files simplified as requested.
erDiagram
  %% Relationships
  Teams ||--o{ TeamMembers : has
  Teams ||--o{ Seasons : has
  Teams ||--o{ TeamGoals : owns
  Teams ||--o{ Invites : sends
  Teams ||--o{ TeamSettings : config

  Seasons ||--o{ TeamGoals : contains
  Seasons ||--o{ MemberGoals : contains
  Seasons ||--o{ Progress : aggregates

  TeamGoals ||--o{ Progress : trackedBy
  MemberGoals ||--o{ Progress : trackedBy

  TeamGoals ||--o{ Comments : discussedBy
  MemberGoals ||--o{ Comments : discussedBy
  Progress ||--o{ Comments : discussedBy

  Comments ||--o{ Files : has

  %% Tables
  Teams {
    string teamId PK "UUID"
    string name
    string status "active|inactive"
    datetime createdAt
    datetime updatedAt
    datetime deletedAt "nullable"
  }

  TeamSettings {
    string teamId PK "FK to Teams"
    string timezone "e.g., Europe/Zurich"
    string locale "e.g., en-GB"
    boolean allowFileUploads
    boolean allowComments
    datetime createdAt
    datetime updatedAt
  }

  TeamMembers {
    string teamId PK "FK Teams"
    string cognitoSub PK "Cognito user sub"
    string role "owner|admin|coach|member"
    string status "active|removed|pending"
    string displayName "optional cache"
    datetime joinedAt
    datetime leftAt "nullable"
    datetime createdAt
    datetime updatedAt
  }

  Invites {
    string inviteId PK "UUID"
    string teamId "FK Teams"
    string email
    string role "owner|admin|coach|member"
    string status "pending|accepted|expired|revoked"
    string token "unique"
    string message "optional"
    datetime expiresAt
    datetime acceptedAt "nullable"
    string invitedByCognitoSub
    string acceptedByCognitoSub "nullable"
    datetime createdAt
  }

  Seasons {
    string seasonId PK "UUID"
    string teamId "FK Teams"
    string name
    date startDate
    date endDate
    string status "planned|active|completed|archived"
    datetime createdAt
    datetime updatedAt
  }

  TeamGoals {
    string teamGoalId PK "UUID"
    string teamId "FK Teams"
    string seasonId "FK Seasons"
    string title
    string description
    string status "open|in_progress|done|archived"
    string createdByCognitoSub
    datetime createdAt
    datetime updatedAt
  }

  MemberGoals {
    string memberGoalId PK "UUID"
    string teamId "FK Teams"
    string seasonId "FK Seasons"
    string ownerCognitoSub "goal owner (Cognito sub)"
    string title
    string description
    string status "open|in_progress|done|archived"
    string createdByCognitoSub
    datetime createdAt
    datetime updatedAt
  }

  %% Progress is tied to exactly one goal (team or member) and includes a 0..5 rating
  Progress {
    string progressId PK "UUID"
    string teamId "FK Teams"
    string seasonId "FK Seasons"
    string teamGoalId "FK TeamGoals, nullable"
    string memberGoalId "FK MemberGoals, nullable"
    string authorCognitoSub "who wrote it (Cognito sub)"
    string summary "short title"
    text details "long narrative"
    int rating "0..5"
    datetime createdAt
    datetime updatedAt
  }

  %% Comments attach to exactly one of TeamGoal, MemberGoal, or Progress
  Comments {
    string commentId PK "UUID"
    string teamId "FK Teams"
    string authorCognitoSub "Cognito sub"
    string teamGoalId "FK TeamGoals, nullable"
    string memberGoalId "FK MemberGoals, nullable"
    string progressId "FK Progress, nullable"
    string content
    datetime createdAt
  }

  %% Files only attach to comments
  Files {
    string fileId PK "UUID"
    string commentId "FK Comments"
    string storageKey "path/key in storage"
    string filename
    datetime createdAt
  }
```
## User Invite Flow
```mermaid
flowchart TD
  A[Inviter creates invite] --> B[Backend: create Invite record + token_hash]
  B --> C[Send single app email with accept link]
  C --> D[Invitee clicks link => preview page]
  D --> E{Is there existing Cognito user?}
  E -->|Yes| F[Attach existing user to Team create TeamMember]
  E -->|No| G[User sets password -> AdminSetUserPassword; set email_verified]
  F --> H[Mark invite accepted, return success]
  G --> H
```

### Detailed
```mermaid
sequenceDiagram
  autonumber
  participant Inviter
  participant Backend
  participant DB as Database
  participant Cognito
  participant Email
  participant Frontend
  participant Invitee

  Note over Backend,DB: Invite creation
  Inviter->>Backend: POST /teams/{teamId}/invites {email, role, message}
  Backend->>DB: create Invite {id,email,role,token_hash,expiresAt,status=pending}
  alt create Cognito user at invite time (opt)
    Backend->>Cognito: AdminCreateUser(email, MessageAction=SUPPRESS)
  end
  Backend->>Email: Send single app-controlled email (link: https://app/accept?token=<raw-token>)
  Backend->>Inviter: 201 {inviteId, expiresAt}

  Note over Invitee,Frontend: Invitee clicks link
  Invitee->>Frontend: GET /accept?token=<raw-token>
  Frontend->>Backend: GET /invites/preview?token=<raw-token>  -- show invite details
  Backend->>DB: find invite by token_hash, validate pending/not expired
  Backend-->>Frontend: 200 {teamName, invitedBy, role, expiresAt}

  Note over Frontend: User chooses Accept
  alt invitee is already signed-in as the same email
    Frontend->>Backend: POST /invites/complete {token} with Authorization
    Backend->>DB: validate invite, find invite.email
    Backend->>Cognito: AdminGetUser(email) -> get sub
    Backend->>DB: create TeamMember {cognitoSub, teamId, role, status=active}
    Backend->>DB: mark invite accepted
    Backend-->>Frontend: 200 {teamMember, redirect}
  else invitee is not signed-in or is new
    Frontend->>Backend: POST /invites/complete {token, password, displayName?}
    Backend->>DB: validate invite, find invite.email
    Backend->>Cognito: AdminGetUser/ensure user exists -> get cognito username
    Backend->>Cognito: AdminSetUserPassword(username, password, Permanent=true)
    Backend->>Cognito: AdminUpdateUserAttributes(username, email_verified=true)
    Backend->>DB: create TeamMember {cognitoSub, teamId, role, status=active}
    Backend->>DB: mark invite accepted
    Backend-->>Frontend: 200 {teamMember, tokens? or redirect}
  end

  Note over Backend,DB: Invite consumed (single-use)
```