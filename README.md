# RBAC System

A simplified Role-Based Access Control (RBAC) system built with Go and MongoDB featuring multi-tenancy, organization management, and third-party authentication.

## Overview

This RBAC system provides a lightweight alternative to more complex systems like Keycloak, focusing on core identity and access management capabilities. It's designed to be easily integrable with web applications while maintaining a small footprint.

## Features

- **User Management**: Create, read, update, and delete users
- **Role-Based Access Control**: Define and manage roles with specific permissions
- **Organization Multi-tenancy**: Support for multiple organizations with isolated resources
- **Authentication**: 
  - Local username/password authentication
  - Third-party OAuth (Google, GitHub)
- **JWT-based Authorization**: Securely manage sessions with JSON Web Tokens
- **RESTful API**: Modern API design following REST principles
- **React Frontend**: Clean and intuitive user interface built with React and Tailwind CSS

## Technology Stack

### Backend
- **Language**: Go 1.23+
- **Web Framework**: Fiber
- **Database**: MongoDB
- **Authentication**: JWT, OAuth2 (via Goth library)
- **Documentation**: OpenAPI (Swagger)

### Frontend
- **Framework**: React with TypeScript
- **State Management**: React Context API
- **Styling**: Tailwind CSS
- **HTTP Client**: Axios
- **Routing**: React Router

## Project Structure

```
.
├── config/                 # Application configuration
├── database/               # Database connection and utilities
├── frontend/               # React frontend application
│   ├── public/             # Static assets
│   └── src/                # React source code
│       ├── components/     # Reusable UI components
│       ├── context/        # React context providers
│       ├── pages/          # Page components
│       ├── services/       # API services
│       └── types/          # TypeScript type definitions
├── handlers/               # HTTP request handlers
├── middleware/             # HTTP middleware
├── models/                 # Data models
├── routes/                 # API route definitions
├── scripts/                # Utility scripts
├── services/               # Business logic
└── utils/                  # Helper functions
```

## Getting Started

### Prerequisites

- Go 1.23 or later
- Node.js 18 or later
- MongoDB instance (local or Atlas)

### Environment Variables

Create a `.env` file in the root directory with the following variables:

```
PORT=5000
MONGO_URI=your_mongodb_connection_string
DATABASE_NAME=rbac_system
JWT_SECRET=your_secure_jwt_secret_key
JWT_EXPIRATION_HOURS=24
CORS_ALLOW_ORIGINS=*

# OAuth2 credentials (optional)
GOOGLE_CLIENT_ID=your_google_client_id
GOOGLE_CLIENT_SECRET=your_google_client_secret
GITHUB_CLIENT_ID=your_github_client_id
GITHUB_CLIENT_SECRET=your_github_client_secret
```

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/rbac-system.git
   cd rbac-system
   ```

2. Install backend dependencies:
   ```bash
   go mod download
   ```

3. Install frontend dependencies:
   ```bash
   cd frontend
   npm install
   cd ..
   ```

4. Build the frontend:
   ```bash
   cd frontend
   npm run build
   cd ..
   ```

5. Start the application:
   ```bash
   go run main.go
   ```

The application will be accessible at http://localhost:5000.

## Documentation

For detailed documentation, please check the following:

- [API Documentation](docs/API.md) - Detailed API reference for all endpoints
- [Architecture Guide](docs/ARCHITECTURE.md) - System design and component architecture
- [Integration Guide](docs/INTEGRATION.md) - How to integrate with other applications
- [Contributing Guide](docs/CONTRIBUTING.md) - Guidelines for contributing to the project

The above documentation provides comprehensive details about the system. Below is a brief overview of the available API endpoints.

## Data Models

### User

```json
{
  "id": "string",
  "username": "string",
  "email": "string",
  "firstName": "string",
  "lastName": "string",
  "active": "boolean",
  "emailVerified": "boolean",
  "roleIds": ["string"],
  "organizationIds": ["string"],
  "authProvider": "string",
  "createdAt": "string",
  "updatedAt": "string",
  "lastLogin": "string"
}
```

### Role

```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "organizationId": "string",
  "permissionIds": ["string"],
  "isSystemDefault": "boolean",
  "createdAt": "string",
  "updatedAt": "string"
}
```

### Permission

```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "resource": "string",
  "action": "string",
  "organizationId": "string",
  "isSystemDefault": "boolean",
  "createdAt": "string",
  "updatedAt": "string"
}
```

### Organization

```json
{
  "id": "string",
  "name": "string",
  "description": "string",
  "domain": "string",
  "active": "boolean",
  "adminIds": ["string"],
  "createdAt": "string",
  "updatedAt": "string"
}
```

## Security Considerations

- All endpoints (except authentication) require a valid JWT token.
- Passwords are hashed using bcrypt before storage.
- Role-based authorization limits access to sensitive operations.
- API requests are validated to prevent injection attacks.
- Cross-Origin Resource Sharing (CORS) is configured to restrict unauthorized domains.

## Development

### Frontend Development

Start the frontend development server:

```bash
cd frontend
npm start
```

This will start the React development server on port 3000 with hot reloading.

### Backend Development

For hot reloading during backend development, you can use tools like Air:

```bash
air
```

### Testing

Run backend tests:

```bash
go test ./...
```

Run frontend tests:

```bash
cd frontend
npm test
```

## Deployment

The application can be deployed as a single binary with the frontend assets embedded.

1. Build the frontend:
   ```bash
   cd frontend
   npm run build
   cd ..
   ```

2. Build the Go application:
   ```bash
   go build -o rbac-system
   ```

3. Deploy the binary along with your `.env` file to your server.

## Extending the System

### Adding Custom Permissions

1. Define new permission constants in `models/permission.go`
2. Create migration scripts to add these permissions to the database
3. Update relevant handlers to check for these permissions

### Adding Third-Party Authentication Providers

1. Add the provider's configuration to the `.env` file
2. Import the provider from the Goth library in `handlers/auth_handler.go`
3. Add the provider to the `goth.UseProviders()` call in the `NewAuthHandler` function
4. Create new handler methods for the OAuth flow

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgements

- [Fiber](https://github.com/gofiber/fiber) - Express-inspired web framework for Go
- [Goth](https://github.com/markbates/goth) - Multi-provider authentication for Go
- [MongoDB Go Driver](https://github.com/mongodb/mongo-go-driver) - Official MongoDB driver for Go
- [React](https://reactjs.org/) - JavaScript library for building user interfaces
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework