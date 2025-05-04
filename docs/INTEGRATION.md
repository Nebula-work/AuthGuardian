# Integrating with RBAC System

This guide explains how to integrate the RBAC System with your applications for authentication and authorization.

## Table of Contents

- [Integration Options](#integration-options)
- [REST API Integration](#rest-api-integration)
- [JWT Verification](#jwt-verification)
- [OAuth Integration](#oauth-integration)
- [User Provisioning](#user-provisioning)
- [Sample Code](#sample-code)

## Integration Options

The RBAC System offers several integration options:

1. **REST API Integration**: Use the RBAC API directly from your application
2. **JWT Verification**: Verify JWT tokens issued by the RBAC System
3. **OAuth Delegation**: Use the RBAC System as an OAuth provider
4. **User Provisioning**: Automatically create and manage users in the RBAC System

## REST API Integration

The most direct integration is to use the RBAC System's REST API for authentication and authorization.

### Authentication Flow

1. Direct users to the RBAC System's login page or implement the login form in your application
2. On successful login, the RBAC System returns a JWT token
3. Store this token securely in your application (e.g., in memory, secure cookies, or local storage)
4. Include the token in all subsequent API requests to the RBAC System

### API Request Format

Include the JWT token in the Authorization header:

```http
GET /api/users
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### Error Handling

Handle authentication and authorization errors from the API:

- `401 Unauthorized`: Token is missing, invalid, or expired
- `403 Forbidden`: User doesn't have the required permissions
- `404 Not Found`: Resource doesn't exist
- `500 Internal Server Error`: Server-side error

### Sample Integration Process

1. Implement login functionality:
   ```javascript
   async function login(username, password) {
     const response = await fetch('http://rbac-system-url/api/auth/login', {
       method: 'POST',
       headers: { 'Content-Type': 'application/json' },
       body: JSON.stringify({ username, password })
     });
     
     const data = await response.json();
     if (!data.success) {
       throw new Error(data.error);
     }
     
     // Store the token
     localStorage.setItem('rbac_token', data.token);
     
     return data.user;
   }
   ```

2. Make authenticated requests:
   ```javascript
   async function fetchWithAuth(url, options = {}) {
     const token = localStorage.getItem('rbac_token');
     
     const response = await fetch(url, {
       ...options,
       headers: {
         ...options.headers,
         'Authorization': `Bearer ${token}`
       }
     });
     
     if (response.status === 401) {
       // Token expired or invalid, redirect to login
       redirectToLogin();
       return null;
     }
     
     return response;
   }
   ```

## JWT Verification

If your application needs to verify JWT tokens issued by the RBAC System independently, you can implement token verification using the shared secret.

### Token Structure

The JWT token contains the following claims:

- `sub`: User ID
- `username`: Username
- `email`: User email
- `roles`: Array of role IDs
- `orgs`: Array of organization IDs
- `exp`: Expiration timestamp
- `iat`: Issued at timestamp

### Verification Process

1. Get the JWT secret from the RBAC System configuration
2. Verify the token signature, expiration, and issuer
3. Extract user information from the token claims
4. Use the role and organization information for authorization decisions

### Sample Verification Code (Node.js)

```javascript
const jwt = require('jsonwebtoken');

function verifyToken(token, secret) {
  try {
    const decoded = jwt.verify(token, secret);
    
    // Check if token is expired
    const now = Math.floor(Date.now() / 1000);
    if (decoded.exp < now) {
      throw new Error('Token expired');
    }
    
    return {
      userId: decoded.sub,
      username: decoded.username,
      email: decoded.email,
      roleIds: decoded.roles,
      organizationIds: decoded.orgs
    };
  } catch (error) {
    throw new Error(`Invalid token: ${error.message}`);
  }
}
```

### Sample Verification Code (Go)

```go
package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type CustomClaims struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	Orgs     []string `json:"orgs"`
	jwt.StandardClaims
}

func verifyToken(tokenString, secret string) (*CustomClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	// Check if token is expired
	if claims.ExpiresAt < time.Now().Unix() {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}
```

## OAuth Integration

The RBAC System supports OAuth 2.0 for third-party authentication. You can integrate with OAuth providers in two ways:

### Using Supported OAuth Providers

The RBAC System already supports Google and GitHub OAuth. To use these providers:

1. Configure the OAuth provider credentials in the RBAC System
2. Direct users to the appropriate OAuth endpoint:
   - Google: `/api/auth/oauth/google`
   - GitHub: `/api/auth/oauth/github`
3. The RBAC System will handle the OAuth flow and return a JWT token

### Adding New OAuth Providers

To add a new OAuth provider:

1. Register an application with the OAuth provider
2. Add the provider credentials to the RBAC System configuration
3. Implement the provider-specific logic in the RBAC System
4. Create new endpoints for the OAuth flow

## User Provisioning

If your application needs to create and manage users in the RBAC System, you can use the user management API endpoints.

### Creating Users

```javascript
async function createUser(userData) {
  const response = await fetchWithAuth('http://rbac-system-url/api/users', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(userData)
  });
  
  const data = await response.json();
  if (!data.success) {
    throw new Error(data.error);
  }
  
  return data.data;
}
```

### Assigning Roles

```javascript
async function assignRolesToUser(userId, roleIds) {
  const user = await fetchUser(userId);
  
  // Update user with new roles
  const updatedUser = {
    ...user,
    roleIds: [...new Set([...user.roleIds, ...roleIds])]
  };
  
  const response = await fetchWithAuth(`http://rbac-system-url/api/users/${userId}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(updatedUser)
  });
  
  const data = await response.json();
  if (!data.success) {
    throw new Error(data.error);
  }
  
  return data.data;
}
```

## Sample Code

### Complete Integration Example (JavaScript/React)

```javascript
// auth.js
import { createContext, useContext, useState, useEffect } from 'react';

const API_URL = 'http://rbac-system-url/api';
const TOKEN_KEY = 'rbac_token';
const USER_KEY = 'rbac_user';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  
  useEffect(() => {
    // Check for existing token and user data
    const token = localStorage.getItem(TOKEN_KEY);
    const userData = localStorage.getItem(USER_KEY);
    
    if (token && userData) {
      try {
        setUser(JSON.parse(userData));
      } catch (err) {
        console.error('Failed to parse user data');
      }
    }
    
    setLoading(false);
  }, []);
  
  const login = async (username, password) => {
    try {
      setLoading(true);
      setError(null);
      
      const response = await fetch(`${API_URL}/auth/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });
      
      const data = await response.json();
      
      if (!data.success) {
        throw new Error(data.error || 'Login failed');
      }
      
      localStorage.setItem(TOKEN_KEY, data.token);
      localStorage.setItem(USER_KEY, JSON.stringify(data.user));
      
      setUser(data.user);
      return data.user;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };
  
  const logout = () => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    setUser(null);
  };
  
  const hasPermission = (resource, action) => {
    if (!user || !user.roleIds || user.roleIds.length === 0) {
      return false;
    }
    
    // In a real application, you would check against the permissions
    // associated with the user's roles. This is a simplified example.
    return true;
  };
  
  const authFetch = async (url, options = {}) => {
    const token = localStorage.getItem(TOKEN_KEY);
    
    if (!token) {
      throw new Error('No authentication token found');
    }
    
    const response = await fetch(url, {
      ...options,
      headers: {
        ...options.headers,
        'Authorization': `Bearer ${token}`
      }
    });
    
    if (response.status === 401) {
      // Token expired or invalid
      logout();
      throw new Error('Session expired. Please login again.');
    }
    
    return response;
  };
  
  return (
    <AuthContext.Provider value={{
      user,
      loading,
      error,
      login,
      logout,
      hasPermission,
      authFetch
    }}>
      {children}
    </AuthContext.Provider>
  );
}

export const useAuth = () => useContext(AuthContext);
```

### Usage in a React Component

```jsx
import { useAuth } from './auth';

function LoginPage() {
  const { login, error, loading } = useAuth();
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  
  const handleSubmit = async (e) => {
    e.preventDefault();
    
    try {
      await login(username, password);
      // Redirect on success
      window.location.href = '/dashboard';
    } catch (err) {
      // Error is handled by the auth context
      console.error('Login failed:', err);
    }
  };
  
  return (
    <div>
      <h1>Login</h1>
      {error && <div className="error">{error}</div>}
      <form onSubmit={handleSubmit}>
        <div>
          <label>Username:</label>
          <input
            type="text"
            value={username}
            onChange={(e) => setUsername(e.target.value)}
            required
          />
        </div>
        <div>
          <label>Password:</label>
          <input
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </div>
        <button type="submit" disabled={loading}>
          {loading ? 'Logging in...' : 'Login'}
        </button>
      </form>
      <div>
        <a href={`${API_URL}/auth/oauth/google`}>Login with Google</a>
        <a href={`${API_URL}/auth/oauth/github`}>Login with GitHub</a>
      </div>
    </div>
  );
}

function ProtectedRoute({ children, requiredPermission }) {
  const { user, loading, hasPermission } = useAuth();
  
  if (loading) {
    return <div>Loading...</div>;
  }
  
  if (!user) {
    return <Navigate to="/login" />;
  }
  
  if (requiredPermission) {
    const [resource, action] = requiredPermission.split(':');
    if (!hasPermission(resource, action)) {
      return <div>Access Denied</div>;
    }
  }
  
  return children;
}

function App() {
  return (
    <AuthProvider>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/dashboard" element={
          <ProtectedRoute>
            <Dashboard />
          </ProtectedRoute>
        } />
        <Route path="/users" element={
          <ProtectedRoute requiredPermission="users:read">
            <UsersList />
          </ProtectedRoute>
        } />
      </Routes>
    </AuthProvider>
  );
}
```

This example demonstrates a complete integration of the RBAC System with a React application, including authentication, authorization, and protected routes.