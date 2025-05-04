# Contributing to RBAC System

Thank you for your interest in contributing to the RBAC System! This document provides guidelines and instructions for contributing to this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Workflow](#development-workflow)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Testing Guidelines](#testing-guidelines)
- [Documentation](#documentation)
- [Issue Reporting](#issue-reporting)

## Code of Conduct

This project adheres to a Code of Conduct that sets expectations for participation in the community. By participating, you are expected to uphold this code. Please report unacceptable behavior to the project maintainers.

We expect all contributors to:

- Be respectful and inclusive in communications
- Be collaborative and open to different viewpoints
- Gracefully accept constructive criticism
- Focus on what is best for the community
- Show empathy towards other community members

## Getting Started

### Prerequisites

To contribute to the RBAC System, you'll need:

- Go 1.23 or later
- Node.js 18 or later
- MongoDB (local instance or connection to Atlas)
- Git

### Setting Up Development Environment

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/YOUR-USERNAME/rbac-system.git
   cd rbac-system
   ```

3. Add the original repository as an upstream remote:
   ```bash
   git remote add upstream https://github.com/ORIGINAL-OWNER/rbac-system.git
   ```

4. Install backend dependencies:
   ```bash
   go mod download
   ```

5. Install frontend dependencies:
   ```bash
   cd frontend
   npm install
   cd ..
   ```

6. Copy the `.env.example` file to `.env` and configure the environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your preferred editor
   ```

## Development Workflow

### Branches

- `main` - The production branch containing stable code
- `develop` - The development branch for integrating features
- `feature/XXX` - Feature branches for new functionality
- `bugfix/XXX` - Bugfix branches for fixing issues
- `docs/XXX` - Documentation branches for documentation updates

### Working on Features or Fixes

1. Sync your fork with the upstream repository:
   ```bash
   git checkout main
   git pull upstream main
   git push origin main
   ```

2. Create a new branch for your feature or fix:
   ```bash
   git checkout -b feature/your-feature-name
   # or
   git checkout -b bugfix/your-bugfix-name
   ```

3. Make your changes, following the [coding standards](#coding-standards)

4. Commit your changes with descriptive messages:
   ```bash
   git commit -m "feat: Add new permission validation"
   # or
   git commit -m "fix: Resolve token validation issue"
   ```

5. Push your branch to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

6. Create a pull request from your branch to the `develop` branch of the original repository

### Running Local Development Servers

#### Backend

```bash
go run main.go
```

#### Frontend

```bash
cd frontend
npm start
```

This will start the frontend development server on port 3000.

## Pull Request Process

1. Ensure your code follows the project's [coding standards](#coding-standards)
2. Update documentation as needed
3. Include tests for new functionality
4. Ensure all tests pass
5. Update the CHANGELOG.md with details of changes if applicable
6. The PR should target the `develop` branch, not `main`
7. PR titles should follow the [Conventional Commits](https://www.conventionalcommits.org/) format:
   - `feat: Add new feature X`
   - `fix: Resolve issue with Y`
   - `docs: Update API documentation`
   - `test: Add tests for feature Z`
   - `refactor: Improve code organization in module W`

## Coding Standards

### Go Code Standards

- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` to format your code
- Follow the package structure of the project
- Add comments to exported functions, types, and constants
- Keep functions focused on a single responsibility
- Write meaningful error messages
- Use context for cancellation and timeouts

### TypeScript/React Standards

- Follow the [Airbnb JavaScript Style Guide](https://github.com/airbnb/javascript)
- Use TypeScript interfaces for prop types and state
- Use functional components with hooks instead of class components
- Keep components focused on a single responsibility
- Use consistent naming conventions
- Follow the file structure of the project
- Format code with Prettier

### Commit Message Guidelines

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types include:
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: Code changes that neither fix a bug nor add a feature
- `test`: Adding or correcting tests
- `chore`: Changes to the build process or auxiliary tools

## Testing Guidelines

### Backend Testing

- Write unit tests for services and utilities
- Write integration tests for handlers
- Aim for good test coverage, especially for critical components
- Use table-driven tests where appropriate
- Run tests with `go test ./...`

### Frontend Testing

- Write unit tests for components and utilities using Jest and React Testing Library
- Write integration tests for complex user flows
- Test both successful and error paths
- Run tests with `npm test`

## Documentation

Documentation is just as important as code. Please update the documentation when you make changes:

1. **API Documentation**: Update `docs/API.md` when changing or adding API endpoints
2. **Architecture Documentation**: Update `docs/ARCHITECTURE.md` when changing the system design
3. **README.md**: Update for significant changes to setup or usage
4. **Code Comments**: Add clear comments for complex logic or non-obvious behavior

## Issue Reporting

When reporting issues, please use the issue templates provided in the repository. Include:

1. A clear, descriptive title
2. Steps to reproduce the issue
3. Expected behavior
4. Actual behavior
5. Screenshots or logs if applicable
6. Environment information:
   - Operating system
   - Go version
   - Node.js version
   - Browser version (for frontend issues)
   - Any other relevant details

Thank you for contributing to the RBAC System!