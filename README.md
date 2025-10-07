# Chirpy

A Twitter-like social media backend API built with Go, featuring user authentication, post management, and premium subscriptions.

## Features

- **User Management**
  - User registration with secure password hashing
  - User authentication with JWT access tokens
  - Refresh token system for extended sessions
  - User profile updates

- **Chirps (Posts)**
  - Create chirps (max 140 characters)
  - Retrieve all chirps with sorting (ascending/descending)
  - Filter chirps by author
  - Delete your own chirps
  - Built-in profanity filter

- **Premium Features**
  - Chirpy Red subscription via Polka webhooks
  - User upgrade system

- **Security**
  - JWT-based authentication
  - Bcrypt password hashing
  - Bearer token authorization
  - API key authentication for webhooks

## Tech Stack

- **Language**: Go
- **Database**: PostgreSQL (via sqlc)
- **Authentication**: JWT tokens, bcrypt
- **Dependencies**:
  - `github.com/google/uuid` - UUID generation
  - Custom auth package for security

## API Endpoints

### Health Check
- `GET /api/healthz` - Service health check

### Users
- `POST /api/users` - Register a new user
- `PUT /api/users` - Update user profile (requires auth)
- `POST /api/login` - Login and receive tokens
- `POST /api/refresh` - Refresh access token
- `POST /api/revoke` - Revoke refresh token

### Chirps
- `POST /api/chirps` - Create a new chirp (requires auth)
- `GET /api/chirps` - Get all chirps (supports `?sort=asc|desc` and `?author_id=<uuid>`)
- `GET /api/chirps/{chirpID}` - Get a specific chirp
- `DELETE /api/chirps/{chirpID}` - Delete your chirp (requires auth)

### Webhooks
- `POST /api/polka/webhooks` - Handle Polka payment webhooks (requires API key)

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL database
- Environment variables configured

### Environment Variables

```bash
JWT_SECRET=your-jwt-secret
POLKA_KEY=your-polka-api-key
DATABASE_URL=your-database-connection-string
```

### Installation

```bash
# Clone the repository
git clone https://github.com/debobrad579/chirpy.git
cd chirpy

# Install dependencies
go mod download

# Run the application
go run .
```

## Project Structure

```
chirpy/
├── internal/
│   ├── auth/          # Authentication utilities
│   └── database/      # Database queries and models (generated using sqlc)
├── main.go            # Application entry point
└── README.md
```

## Authentication

The API uses JWT tokens for authentication. After logging in, you'll receive:
- **Access Token**: Short-lived token (1 hour) for API requests
- **Refresh Token**: Long-lived token (60 days) for obtaining new access tokens

Include the access token in requests:
```
Authorization: Bearer <your-access-token>
```

## License
This project is part of the Boot.dev curriculum.
