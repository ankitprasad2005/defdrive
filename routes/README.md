# API Endpoints Documentation

## User Endpoints

### POST /api/signup

**Request:**
```json
{
  "name": "John Doe",
  "email": "john.doe@example.com",
  "username": "johndoe",
  "password": "password123"
}
```

**Response:**
- Success (200):
  ```json
  {
    "message": "User created successfully",
    "user": {
      "ID": 1,
      "Name": "John Doe",
      "Email": "john.doe@example.com",
      "Username": "johndoe"
    }
  }
  ```
- Error (400):
  ```json
  {
    "error": "Validation error message"
  }
  ```
- Error (500):
  ```json
  {
    "error": "Failed to create user: error message"
  }
  ```

### POST /api/login

**Request:**
```json
{
  "username": "johndoe",
  "password": "password123"
}
```

**Response:**
- Success (200):
  ```json
  {
    "message": "Login successful",
    "token": "jwt_token",
    "user": {
      "id": 1,
      "username": "johndoe",
      "email": "john.doe@example.com",
      "name": "John Doe"
    }
  }
  ```
- Error (400):
  ```json
  {
    "error": "Validation error message"
  }
  ```
- Error (401):
  ```json
  {
    "error": "Invalid username or password"
  }
  ```
- Error (500):
  ```json
  {
    "error": "Failed to generate token"
  }
  ```

## File Endpoints

### POST /api/upload

**Request:**
- Form-data with file field

**Response:**
- Success (200):
  ```json
  {
    "message": "File uploaded successfully",
    "file": {
      "ID": 1,
      "Name": "example.txt",
      "Location": "uploads/example.txt",
      "Size": 12345,
      "Public": false,
      "UserID": 1
    }
  }
  ```
- Error (400):
  ```json
  {
    "error": "No file provided"
  }
  ```
- Error (401):
  ```json
  {
    "error": "User not authenticated"
  }
  ```
- Error (500):
  ```json
  {
    "error": "Failed to save file"
  }
  ```

### GET /api/files

**Response:**
- Success (200):
  ```json
  {
    "files": [
      {
        "ID": 1,
        "Name": "example.txt",
        "Location": "uploads/example.txt",
        "Size": 12345,
        "Public": false,
        "UserID": 1
      }
    ]
  }
  ```
- Error (401):
  ```json
  {
    "error": "User not authenticated"
  }
  ```
- Error (500):
  ```json
  {
    "error": "Failed to retrieve files"
  }
  ```

### PUT /api/files/:fileID/access

**Request:**
```json
{
  "public": true
}
```

**Response:**
- Success (200):
  ```json
  {
    "message": "File access updated successfully",
    "file": {
      "ID": 1,
      "Name": "example.txt",
      "Location": "uploads/example.txt",
      "Size": 12345,
      "Public": true,
      "UserID": 1
    }
  }
  ```
- Error (400):
  ```json
  {
    "error": "Invalid request body"
  }
  ```
- Error (401):
  ```json
  {
    "error": "User not authenticated"
  }
  ```
- Error (403):
  ```json
  {
    "error": "You don't have permission to modify this file"
  }
  ```
- Error (404):
  ```json
  {
    "error": "File not found"
  }
  ```
- Error (500):
  ```json
  {
    "error": "Failed to update file"
  }
  ```

### DELETE /api/files/:fileID

**Response:**
- Success (200):
  ```json
  {
    "message": "File deleted successfully"
  }
  ```
- Error (400):
  ```json
  {
    "error": "Invalid file ID"
  }
  ```
- Error (401):
  ```json
  {
    "error": "User not authenticated"
  }
  ```
- Error (403):
  ```json
  {
    "error": "You don't have permission to delete this file"
  }
  ```
- Error (404):
  ```json
  {
    "error": "File not found"
  }
  ```
- Error (500):
  ```json
  {
    "error": "Failed to delete file record"
  }
  ```

## Access Endpoints

### POST /api/files/:fileID/accesses

**Request:**
```json
{
  "name": "Access Name",
  "subnet": "192.168.1.0/24",
  "ip": "192.168.1.1",
  "expires": "2025-12-31T23:59:59Z",
  "public": true
}
```

**Response:**
- Success (200):
  ```json
  {
    "message": "Access created successfully",
    "access": {
      "ID": 1,
      "Name": "Access Name",
      "Link": "unique-access-link",
      "Subnet": "192.168.1.0/24",
      "IP": "192.168.1.1",
      "Expires": "2025-12-31T23:59:59Z",
      "Public": true,
      "FileID": 1
    }
  }
  ```
- Error (400):
  ```json
  {
    "error": "Invalid request body"
  }
  ```
- Error (401):
  ```json
  {
    "error": "User not authenticated"
  }
  ```
- Error (403):
  ```json
  {
    "error": "You don't have permission to create access for this file"
  }
  ```
- Error (404):
  ```json
  {
    "error": "File not found"
  }
  ```
- Error (500):
  ```json
  {
    "error": "Failed to create access record"
  }
  ```

### GET /api/files/:fileID/accesses

**Response:**
- Success (200):
  ```json
  {
    "accesses": [
      {
        "ID": 1,
        "Name": "Access Name",
        "Link": "unique-access-link",
        "Subnet": "192.168.1.0/24",
        "IP": "192.168.1.1",
        "Expires": "2025-12-31T23:59:59Z",
        "Public": true,
        "FileID": 1
      }
    ]
  }
  ```
- Error (401):
  ```json
  {
    "error": "User not authenticated"
  }
  ```
- Error (403):
  ```json
  {
    "error": "You don't have permission to view accesses for this file"
  }
  ```
- Error (404):
  ```json
  {
    "error": "File not found"
  }
  ```
- Error (500):
  ```json
  {
    "error": "Failed to retrieve accesses"
  }
  ```

### PUT /api/accesses/:accessID

**Request:**
```json
{
  "name": "Updated Access Name",
  "subnet": "192.168.2.0/24",
  "ip": "192.168.2.1",
  "expires": "2026-12-31T23:59:59Z",
  "public": false
}
```

**Response:**
- Success (200):
  ```json
  {
    "message": "Access updated successfully",
    "access": {
      "ID": 1,
      "Name": "Updated Access Name",
      "Link": "unique-access-link",
      "Subnet": "192.168.2.0/24",
      "IP": "192.168.2.1",
      "Expires": "2026-12-31T23:59:59Z",
      "Public": false,
      "FileID": 1
    }
  }
  ```
- Error (400):
  ```json
  {
    "error": "Invalid request body"
  }
  ```
- Error (401):
  ```json
  {
    "error": "User not authenticated"
  }
  ```
- Error (403):
  ```json
  {
    "error": "You don't have permission to update this access"
  }
  ```
- Error (404):
  ```json
  {
    "error": "Access record not found"
  }
  ```
- Error (500):
  ```json
  {
    "error": "Failed to update access record"
  }
  ```

### DELETE /api/accesses/:accessID

**Response:**
- Success (200):
  ```json
  {
    "message": "Access deleted successfully"
  }
  ```
- Error (400):
  ```json
  {
    "error": "Invalid access ID"
  }
  ```
- Error (401):
  ```json
  {
    "error": "User not authenticated"
  }
  ```
- Error (403):
  ```json
  {
    "error": "You don't have permission to delete this access"
  }
  ```
- Error (404):
  ```json
  {
    "error": "Access record not found"
  }
  ```
- Error (500):
  ```json
  {
    "error": "Failed to delete access record"
  }
  ```
