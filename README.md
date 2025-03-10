# Defdrive

Defdrive is a project that allows users to create multiple expiry keys based on time and has features like one-time use of the link or specific data traffic allowed.

## Features

- User authentication with JWT
- Create and manage expiry keys
- One-time use links
- Data traffic control

## Setup

1. Clone the repository
2. Create a `.env` file in the root directory. Take `.env.example` for reference.
3. Ensure a PostgreSQL database is running on the specified port and set all necessary environment variables in the `.env` file.
4. Run `go mod tidy` to install dependencies.
5. Run `go run cmd/main.go` to start the server

## Running with Docker Compose

1. Clone the repository
2. Create a `.env` file in the root directory. Take `.env.example` for reference.
3. Run `docker-compose up --build` to start the services

## API Endpoints

- `POST /api/auth/login`: Authenticate a user and return a JWT token.
- `POST /api/keys`: Create a new expiry key.
- `GET /api/keys`: Retrieve all expiry keys.
- `GET /api/keys/:id`: Retrieve a specific expiry key by ID.
- `DELETE /api/keys/:id`: Delete a specific expiry key by ID.
- `POST /api/links`: Create a one-time use link.
- `GET /api/links/:id`: Retrieve a specific link by ID.
- `DELETE /api/links/:id`: Delete a specific link by ID.
