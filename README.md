# Defdrive

Defdrive is a project that allows users to create multiple expiry keys based on time and has features like one-time use of the link, specific data traffic allowed, subnet restriction, public IP restriction, and TTL (Time to Live).

## Features

- User authentication with JWT
- Create and manage expiry keys
- One-time use links
- Data traffic control
- Subnet restriction
- Public IP restriction
- TTL (Time to Live)
- Public access links for files

## Setup

1. Clone the repository.
2. Create a `.env` file in the root directory. Take `.env.example` for reference.
3. Ensure a PostgreSQL database is running on the specified port and set all necessary environment variables in the `.env` file.
4. Run `go mod tidy` to install dependencies.
5. Run `go run main.go` to start the server.

## Running with Docker Compose

1. Clone the repository.
2. Create a `.env` file in the root directory. Take `.env.example` for reference.
3. Run `docker-compose up --build` to start the services.

## API Endpoints

- `POST /api/signup`: Register a new user.
- `POST /api/login`: Authenticate a user and return a JWT token.
- `POST /api/upload`: Upload a file.
- `GET /api/files`: Retrieve all files for the authenticated user.
- `PUT /api/files/:fileID/access`: Update the public access status of a file.
- `DELETE /api/files/:fileID`: Delete a file.
- `POST /api/files/:fileID/accesses`: Create a new access record for a file.
- `GET /api/files/:fileID/accesses`: Retrieve all access records for a file.
- `PUT /api/accesses/:accessID/access`: Update an access record.
- `DELETE /api/accesses/:accessID`: Delete an access record.
- `GET /link/:hash`: Access a file using a public link.

## Database Models

<p align="center">
 <img src=".github/files/db.png" alt="DB model" />
</p>
