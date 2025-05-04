# API Documentation

This document provides detailed information about the RBAC System API.

## Table of Contents

- [Authentication](#authentication)
- [Users](#users)
- [Roles](#roles)
- [Permissions](#permissions)
- [Organizations](#organizations)
- [Error Handling](#error-handling)

## Base URL

All API endpoints are relative to the base URL:

```
http://localhost:5000/api
```

## Authentication

The API uses JWT (JSON Web Tokens) for authentication. Most endpoints require a valid token.

### Authentication Endpoints

#### Register a New User

```
POST /auth/register
```

Creates a new user account.

**Request Body:**

```json
{
  "username": "johndoe",
  "email": "john.doe@example.com",
  "password": "securePassword123",
  "firstName": "John",
  "lastName": "Doe",
  "organizationIds": ["60d21b4667d0d8992e610c85"]  // Optional
}
```

**Response:**

```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "60d21b4667d0d8992e610c86",
    "username": "johndoe",
    "email": "john.doe@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "active": true,
    "emailVerified": false,
    "roleIds": ["60d21b4667d0d8992e610c87"],
    "organizationIds": ["60d21b4667d0d8992e610c85"],
    "authProvider": "local",
    "createdAt": "2023-05-03T12:00:00Z",
    "updatedAt": "2023-05-03T12:00:00Z"
  }
}
```

#### Login

```
POST /auth/login
```

Authenticates a user and returns a JWT token.

**Request Body:**

```json
{
  "username": "johndoe",
  "password": "securePassword123"
}
```

**Response:**

```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "60d21b4667d0d8992e610c86",
    "username": "johndoe",
    "email": "john.doe@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "active": true,
    "emailVerified": false,
    "roleIds": ["60d21b4667d0d8992e610c87"],
    "organizationIds": ["60d21b4667d0d8992e610c85"],
    "authProvider": "local",
    "createdAt": "2023-05-03T12:00:00Z",
    "updatedAt": "2023-05-03T12:00:00Z",
    "lastLogin": "2023-05-03T15:30:00Z"
  }
}
```

#### Refresh Token

```
POST /auth/refresh-token
```

Refreshes an existing JWT token.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**

```json
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "60d21b4667d0d8992e610c86",
    "username": "johndoe",
    "email": "john.doe@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "active": true,
    "emailVerified": false,
    "roleIds": ["60d21b4667d0d8992e610c87"],
    "organizationIds": ["60d21b4667d0d8992e610c85"],
    "authProvider": "local",
    "createdAt": "2023-05-03T12:00:00Z",
    "updatedAt": "2023-05-03T12:00:00Z",
    "lastLogin": "2023-05-03T15:30:00Z"
  }
}
```

#### OAuth Authentication

##### Google OAuth

```
GET /auth/oauth/google
```

Initiates the Google OAuth flow.

##### Google OAuth Callback

```
GET /auth/oauth/google/callback
```

Handles the Google OAuth callback.

##### GitHub OAuth

```
GET /auth/oauth/github
```

Initiates the GitHub OAuth flow.

##### GitHub OAuth Callback

```
GET /auth/oauth/github/callback
```

Handles the GitHub OAuth callback.

## Users

### User Endpoints

#### Get All Users

```
GET /users
```

Retrieves a list of users with pagination.

**Query Parameters:**

- `limit` (optional, default: 100): Number of users to return
- `skip` (optional, default: 0): Number of users to skip (for pagination)
- `organizationId` (optional): Filter users by organization ID

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**

```json
{
  "success": true,
  "data": {
    "users": [
      {
        "id": "60d21b4667d0d8992e610c86",
        "username": "johndoe",
        "email": "john.doe@example.com",
        "firstName": "John",
        "lastName": "Doe",
        "active": true,
        "emailVerified": false,
        "roleIds": ["60d21b4667d0d8992e610c87"],
        "organizationIds": ["60d21b4667d0d8992e610c85"],
        "authProvider": "local",
        "createdAt": "2023-05-03T12:00:00Z",
        "updatedAt": "2023-05-03T12:00:00Z",
        "lastLogin": "2023-05-03T15:30:00Z"
      },
      // More users...
    ],
    "total": 125,
    "limit": 10,
    "skip": 0
  }
}
```

#### Get User by ID

```
GET /users/:id
```

Retrieves a specific user by ID.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "60d21b4667d0d8992e610c86",
    "username": "johndoe",
    "email": "john.doe@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "active": true,
    "emailVerified": false,
    "roleIds": ["60d21b4667d0d8992e610c87"],
    "organizationIds": ["60d21b4667d0d8992e610c85"],
    "authProvider": "local",
    "createdAt": "2023-05-03T12:00:00Z",
    "updatedAt": "2023-05-03T12:00:00Z",
    "lastLogin": "2023-05-03T15:30:00Z"
  }
}
```

#### Create User

```
POST /users
```

Creates a new user.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:**

```json
{
  "username": "johndoe",
  "email": "john.doe@example.com",
  "password": "securePassword123",
  "firstName": "John",
  "lastName": "Doe",
  "active": true,
  "roleIds": ["60d21b4667d0d8992e610c87"],
  "organizationIds": ["60d21b4667d0d8992e610c85"],
  "authProvider": "local"
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "60d21b4667d0d8992e610c86",
    "username": "johndoe",
    "email": "john.doe@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "active": true,
    "emailVerified": false,
    "roleIds": ["60d21b4667d0d8992e610c87"],
    "organizationIds": ["60d21b4667d0d8992e610c85"],
    "authProvider": "local",
    "createdAt": "2023-05-03T12:00:00Z",
    "updatedAt": "2023-05-03T12:00:00Z"
  }
}
```

#### Update User

```
PUT /users/:id
```

Updates an existing user.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:**

```json
{
  "firstName": "John",
  "lastName": "Doe Updated",
  "active": true,
  "roleIds": ["60d21b4667d0d8992e610c87", "60d21b4667d0d8992e610c88"],
  "organizationIds": ["60d21b4667d0d8992e610c85"]
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "60d21b4667d0d8992e610c86",
    "username": "johndoe",
    "email": "john.doe@example.com",
    "firstName": "John",
    "lastName": "Doe Updated",
    "active": true,
    "emailVerified": false,
    "roleIds": ["60d21b4667d0d8992e610c87", "60d21b4667d0d8992e610c88"],
    "organizationIds": ["60d21b4667d0d8992e610c85"],
    "authProvider": "local",
    "createdAt": "2023-05-03T12:00:00Z",
    "updatedAt": "2023-05-03T15:45:00Z",
    "lastLogin": "2023-05-03T15:30:00Z"
  }
}
```

#### Delete User

```
DELETE /users/:id
```

Deletes a user.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**

```json
{
  "success": true,
  "message": "User deleted successfully"
}
```

## Roles

### Role Endpoints

#### Get All Roles

```
GET /roles
```

Retrieves a list of roles with pagination.

**Query Parameters:**

- `limit` (optional, default: 100): Number of roles to return
- `skip` (optional, default: 0): Number of roles to skip (for pagination)
- `organizationId` (optional): Filter roles by organization ID

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**

```json
{
  "success": true,
  "data": {
    "roles": [
      {
        "id": "60d21b4667d0d8992e610c87",
        "name": "Admin",
        "description": "Administrator role with full access",
        "organizationId": "60d21b4667d0d8992e610c85",
        "permissionIds": ["60d21b4667d0d8992e610c89", "60d21b4667d0d8992e610c90"],
        "isSystemDefault": true,
        "createdAt": "2023-05-03T12:00:00Z",
        "updatedAt": "2023-05-03T12:00:00Z"
      },
      // More roles...
    ],
    "total": 5,
    "limit": 10,
    "skip": 0
  }
}
```

#### Get Role by ID

```
GET /roles/:id
```

Retrieves a specific role by ID.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "60d21b4667d0d8992e610c87",
    "name": "Admin",
    "description": "Administrator role with full access",
    "organizationId": "60d21b4667d0d8992e610c85",
    "permissionIds": ["60d21b4667d0d8992e610c89", "60d21b4667d0d8992e610c90"],
    "isSystemDefault": true,
    "createdAt": "2023-05-03T12:00:00Z",
    "updatedAt": "2023-05-03T12:00:00Z"
  }
}
```

#### Create Role

```
POST /roles
```

Creates a new role.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:**

```json
{
  "name": "Project Manager",
  "description": "Manages projects and teams",
  "organizationId": "60d21b4667d0d8992e610c85",
  "permissionIds": ["60d21b4667d0d8992e610c89"],
  "isSystemDefault": false
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "60d21b4667d0d8992e610c88",
    "name": "Project Manager",
    "description": "Manages projects and teams",
    "organizationId": "60d21b4667d0d8992e610c85",
    "permissionIds": ["60d21b4667d0d8992e610c89"],
    "isSystemDefault": false,
    "createdAt": "2023-05-03T16:00:00Z",
    "updatedAt": "2023-05-03T16:00:00Z"
  }
}
```

#### Update Role

```
PUT /roles/:id
```

Updates an existing role.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:**

```json
{
  "name": "Project Manager",
  "description": "Manages projects, teams, and resources",
  "permissionIds": ["60d21b4667d0d8992e610c89", "60d21b4667d0d8992e610c91"]
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "60d21b4667d0d8992e610c88",
    "name": "Project Manager",
    "description": "Manages projects, teams, and resources",
    "organizationId": "60d21b4667d0d8992e610c85",
    "permissionIds": ["60d21b4667d0d8992e610c89", "60d21b4667d0d8992e610c91"],
    "isSystemDefault": false,
    "createdAt": "2023-05-03T16:00:00Z",
    "updatedAt": "2023-05-03T16:15:00Z"
  }
}
```

#### Delete Role

```
DELETE /roles/:id
```

Deletes a role.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**

```json
{
  "success": true,
  "message": "Role deleted successfully"
}
```

#### Update Role Permissions

```
PUT /roles/:id/permissions
```

Updates the permissions assigned to a role.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:**

```json
{
  "permissionIds": ["60d21b4667d0d8992e610c89", "60d21b4667d0d8992e610c90", "60d21b4667d0d8992e610c91"]
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "60d21b4667d0d8992e610c88",
    "name": "Project Manager",
    "description": "Manages projects, teams, and resources",
    "organizationId": "60d21b4667d0d8992e610c85",
    "permissionIds": ["60d21b4667d0d8992e610c89", "60d21b4667d0d8992e610c90", "60d21b4667d0d8992e610c91"],
    "isSystemDefault": false,
    "createdAt": "2023-05-03T16:00:00Z",
    "updatedAt": "2023-05-03T16:30:00Z"
  }
}
```

## Permissions

### Permission Endpoints

#### Get All Permissions

```
GET /permissions
```

Retrieves a list of permissions with pagination.

**Query Parameters:**

- `limit` (optional, default: 100): Number of permissions to return
- `skip` (optional, default: 0): Number of permissions to skip (for pagination)
- `organizationId` (optional): Filter permissions by organization ID
- `resource` (optional): Filter permissions by resource
- `action` (optional): Filter permissions by action

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**

```json
{
  "success": true,
  "data": {
    "permissions": [
      {
        "id": "60d21b4667d0d8992e610c89",
        "name": "View Users",
        "description": "Can view user information",
        "resource": "users",
        "action": "read",
        "organizationId": null,
        "isSystemDefault": true,
        "createdAt": "2023-05-03T12:00:00Z",
        "updatedAt": "2023-05-03T12:00:00Z"
      },
      // More permissions...
    ],
    "total": 15,
    "limit": 10,
    "skip": 0
  }
}
```

#### Get Permission by ID

```
GET /permissions/:id
```

Retrieves a specific permission by ID.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "60d21b4667d0d8992e610c89",
    "name": "View Users",
    "description": "Can view user information",
    "resource": "users",
    "action": "read",
    "organizationId": null,
    "isSystemDefault": true,
    "createdAt": "2023-05-03T12:00:00Z",
    "updatedAt": "2023-05-03T12:00:00Z"
  }
}
```

#### Create Permission

```
POST /permissions
```

Creates a new permission.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:**

```json
{
  "name": "Delete Projects",
  "description": "Can delete projects",
  "resource": "projects",
  "action": "delete",
  "organizationId": "60d21b4667d0d8992e610c85",
  "isSystemDefault": false
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "60d21b4667d0d8992e610c92",
    "name": "Delete Projects",
    "description": "Can delete projects",
    "resource": "projects",
    "action": "delete",
    "organizationId": "60d21b4667d0d8992e610c85",
    "isSystemDefault": false,
    "createdAt": "2023-05-03T17:00:00Z",
    "updatedAt": "2023-05-03T17:00:00Z"
  }
}
```

#### Update Permission

```
PUT /permissions/:id
```

Updates an existing permission.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:**

```json
{
  "name": "Delete Projects",
  "description": "Can delete projects and associated resources",
  "resource": "projects",
  "action": "delete"
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "60d21b4667d0d8992e610c92",
    "name": "Delete Projects",
    "description": "Can delete projects and associated resources",
    "resource": "projects",
    "action": "delete",
    "organizationId": "60d21b4667d0d8992e610c85",
    "isSystemDefault": false,
    "createdAt": "2023-05-03T17:00:00Z",
    "updatedAt": "2023-05-03T17:15:00Z"
  }
}
```

#### Delete Permission

```
DELETE /permissions/:id
```

Deletes a permission.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**

```json
{
  "success": true,
  "message": "Permission deleted successfully"
}
```

## Organizations

### Organization Endpoints

#### Get All Organizations

```
GET /organizations
```

Retrieves a list of organizations with pagination.

**Query Parameters:**

- `limit` (optional, default: 100): Number of organizations to return
- `skip` (optional, default: 0): Number of organizations to skip (for pagination)

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**

```json
{
  "success": true,
  "data": {
    "organizations": [
      {
        "id": "60d21b4667d0d8992e610c85",
        "name": "Acme Corp",
        "description": "Leading widgets manufacturer",
        "domain": "acme.com",
        "active": true,
        "adminIds": ["60d21b4667d0d8992e610c86"],
        "createdAt": "2023-05-03T12:00:00Z",
        "updatedAt": "2023-05-03T12:00:00Z"
      },
      // More organizations...
    ],
    "total": 3,
    "limit": 10,
    "skip": 0
  }
}
```

#### Get Organization by ID

```
GET /organizations/:id
```

Retrieves a specific organization by ID.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "60d21b4667d0d8992e610c85",
    "name": "Acme Corp",
    "description": "Leading widgets manufacturer",
    "domain": "acme.com",
    "active": true,
    "adminIds": ["60d21b4667d0d8992e610c86"],
    "createdAt": "2023-05-03T12:00:00Z",
    "updatedAt": "2023-05-03T12:00:00Z"
  }
}
```

#### Create Organization

```
POST /organizations
```

Creates a new organization.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:**

```json
{
  "name": "Globex Corporation",
  "description": "International technology company",
  "domain": "globex.com",
  "active": true,
  "adminIds": ["60d21b4667d0d8992e610c86"]
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "60d21b4667d0d8992e610c93",
    "name": "Globex Corporation",
    "description": "International technology company",
    "domain": "globex.com",
    "active": true,
    "adminIds": ["60d21b4667d0d8992e610c86"],
    "createdAt": "2023-05-03T18:00:00Z",
    "updatedAt": "2023-05-03T18:00:00Z"
  }
}
```

#### Update Organization

```
PUT /organizations/:id
```

Updates an existing organization.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Request Body:**

```json
{
  "name": "Globex Corporation",
  "description": "International technology and innovation company",
  "domain": "globex.com",
  "active": true,
  "adminIds": ["60d21b4667d0d8992e610c86", "60d21b4667d0d8992e610c94"]
}
```

**Response:**

```json
{
  "success": true,
  "data": {
    "id": "60d21b4667d0d8992e610c93",
    "name": "Globex Corporation",
    "description": "International technology and innovation company",
    "domain": "globex.com",
    "active": true,
    "adminIds": ["60d21b4667d0d8992e610c86", "60d21b4667d0d8992e610c94"],
    "createdAt": "2023-05-03T18:00:00Z",
    "updatedAt": "2023-05-03T18:15:00Z"
  }
}
```

#### Delete Organization

```
DELETE /organizations/:id
```

Deletes an organization.

**Headers:**

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

**Response:**

```json
{
  "success": true,
  "message": "Organization deleted successfully"
}
```

## Error Handling

The API uses consistent error responses across all endpoints.

### Error Response Format

```json
{
  "success": false,
  "error": "Error message describing what went wrong"
}
```

### Common HTTP Status Codes

- `200 OK`: Request succeeded
- `201 Created`: Resource created successfully
- `400 Bad Request`: Invalid request parameters
- `401 Unauthorized`: Authentication required or failed
- `403 Forbidden`: Permission denied
- `404 Not Found`: Resource not found
- `409 Conflict`: Resource already exists
- `500 Internal Server Error`: Server-side error