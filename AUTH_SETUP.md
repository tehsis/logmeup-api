# Auth0 Authentication Setup for LogMeUp API

This document explains how to set up Auth0 authentication for the LogMeUp API.

## Prerequisites

1. Auth0 account (sign up at https://auth0.com)
2. PostgreSQL database running
3. Go environment set up

## Auth0 Configuration

### 1. Create Auth0 Application

1. Log in to your Auth0 dashboard
2. Go to Applications > Create Application
3. Choose "Single Page Application" for the frontend
4. Note down the Domain and Client ID

### 2. Create Auth0 API

1. Go to APIs > Create API
2. Set the identifier (e.g., `https://api.logmeup.com`)
3. Choose RS256 as the signing algorithm
4. Note down the API identifier (this will be your audience)

### 3. Configure Allowed Origins

In your Auth0 application settings:
- **Allowed Callback URLs**: `http://localhost:3000/callback, http://localhost:5173/callback`
- **Allowed Logout URLs**: `http://localhost:3000, http://localhost:5173`
- **Allowed Web Origins**: `http://localhost:3000, http://localhost:5173`

### 4. Enable Social Connections

1. Go to Authentication > Social
2. Enable Google and GitHub connections
3. Configure with your OAuth app credentials

## Environment Configuration

### Backend (.env)

Update your `.env` file with Auth0 configuration:

```env
# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=logmeup
SERVER_PORT=8080

# Auth0 Configuration
AUTH0_DOMAIN=your-auth0-domain.auth0.com
AUTH0_AUDIENCE=https://api.logmeup.com
```

Replace `your-auth0-domain` with your actual Auth0 domain.

### Frontend

The frontend Auth0 configuration is fetched from the server at runtime via `/api/auth-config` endpoint.

## Database Migration

Run the migration to add user_id fields:

```bash
migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/logmeup?sslmode=disable" up
```

## Testing Authentication

### 1. Start the API Server

```bash
go run cmd/api/main.go
```

### 2. Test Protected Endpoints

All API endpoints under `/api/*` now require authentication:

```bash
# This will return 401 Unauthorized
curl http://localhost:8080/api/notes

# This requires a valid JWT token
curl -H "Authorization: Bearer YOUR_JWT_TOKEN" http://localhost:8080/api/notes
```

### 3. Get JWT Token

You can get a JWT token by:
1. Logging in through the frontend application
2. Using Auth0's test tab in the API settings
3. Using Auth0's authentication API directly

## Security Features

- **JWT Validation**: All tokens are validated using Auth0's public keys
- **User Isolation**: All data is filtered by user_id from the JWT claims
- **Token Caching**: JWKS keys are cached for 1 hour for performance
- **Automatic Refresh**: Frontend automatically refreshes tokens

## API Changes

### Request Headers

All API requests now require:
```
Authorization: Bearer <jwt_token>
```

### Response Changes

All models now include `user_id` field:

```json
{
  "id": 1,
  "user_id": "auth0|123456789",
  "content": "My note",
  "date": "2024-05-27",
  "created_at": "2024-05-27T10:00:00Z",
  "updated_at": "2024-05-27T10:00:00Z"
}
```

## Troubleshooting

### Common Issues

1. **401 Unauthorized**: Check that the JWT token is valid and not expired
2. **Invalid Issuer**: Ensure AUTH0_DOMAIN is correctly set
3. **Invalid Audience**: Ensure AUTH0_AUDIENCE matches your API identifier
4. **CORS Issues**: Check that your frontend URL is in Auth0's allowed origins

### Debug Mode

Enable debug logging by setting log level to debug in your application.

### Token Validation

You can validate JWT tokens at https://jwt.io using your Auth0 public key.

## Production Considerations

1. Use environment-specific Auth0 applications
2. Set up proper CORS policies
3. Use HTTPS in production
4. Monitor Auth0 logs for authentication issues
5. Set up proper error handling for token refresh failures 