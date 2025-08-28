# Digital Book Lending API

A RESTful API service for digital book lending built with Go, Gin framework, and MySQL database. This application helps users to borrow books. Users should be able to browse the catalogue, borrow books, and return them.

## 🚀 Features

- **User Management**
  - User registration and authentication
  - JWT-based authorization
  - Role-based access control
  - Secure password hashing

- **Book Management**
  - Create, read, update, and delete books
  - Book categorization and inventory tracking

- **Lending Management**
  - Borrow and return books
  - Track lending history
  - Rule limit lending books

- **System Features**
  - Database migrations
  - CORS support
  - Request logging and monitoring
  - Health check endpoint
  - Dockerized deployment

## 🏗️ Architecture

The application follows a clean architecture pattern with the following structure:

```
digital-book-lending/
├── app/                   # Application routing and middleware
├── controller/            # HTTP request handlers
├── database/              # Database connection and configuration
├── interfaces/            # Interface definitions
├── middleware/            # Custom middleware functions
├── migrations/            # Database migration files
├── models/                # Data models and structures
├── repository/            # Data access layer
├── utils/                 # Utility functions and helpers
├── config/                # Configuration files
├── main.go                # Application entry point
├── Dockerfile             # Docker configuration
└── entrypoint.sh          # Docker entrypoint script
```

## 🛠️ Technology Stack

- **Language**: Go 1.24
- **Web Framework**: Gin
- **Database**: MySQL
- **ORM**: GORM
- **Authentication**: JWT (golang-jwt/jwt)
- **Migration**: golang-migrate
- **Caching**: Redis (optional)
- **Configuration**: Viper
- **Containerization**: Docker

## 📋 Prerequisites

- Go 1.24 or higher
- MySQL 8.0 or higher
- Redis (optional, for caching)
- Docker (for containerized deployment)

## ⚙️ Installation

### 1. Clone the repository

```bash
git clone https://github.com/zazhedho/digital-book-lending.git
cd digital-book-lending
```

### 2. Install dependencies

```bash
go mod tidy
```

### 3. Environment Configuration

Create a `.env` file in the root directory:

```env
# Application Configuration
APP_NAME=digital-book-lending
APP_ENV=development
PORT=8080

# Database Configuration
DB_HOST=localhost
DB_PORT=3306
DB_USERNAME=your_db_user
DB_PASS=your_db_password
DB_NAME=book_lending_db

# Migration Configuration
PATH_MIGRATE=file://migrations

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key

# Configuration Management
CONFIG_ID=your-config-id
```

### 4. Database Setup

Make sure MySQL is running and create the database:

```sql
CREATE DATABASE book_lending_db;
```

### 5. Run the application

```bash
go run main.go
```

The server will start on the port specified in your `.env` file (default: 8080).

## 🐳 Docker Deployment

### Build the Docker image

```bash
docker build --build-arg SERVICE_NAME=digital-book-lending -t digital-book-lending .
```

### Run with Docker

```bash
docker run -p 8080:8080 --env-file .env digital-book-lending
```

### Using Docker Compose (recommended)

Create a `docker-compose.yml` file:

```yaml
version: '3.8'

services:
  app:
    build:
      context: .
      args:
        SERVICE_NAME: digital-book-lending
    ports:
      - "8080:8080"
    environment:
      - APP_NAME=digital-book-lending
      - APP_ENV=production
      - PORT=8080
      - DB_HOST=mysql
      - DB_PORT=3306
      - DB_USERNAME=root
      - DB_PASS=password
      - DB_NAME=book_lending_db
    depends_on:
      - mysql
      - redis

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: password
      MYSQL_DATABASE: book_lending_db
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  mysql_data:
```

Run with:

```bash
docker-compose up -d
```

## 📚 API Documentation

### Base URL

```
http://localhost:8080/api/v1
```

### Postman Collection

```
https://zaiduszhuhur.postman.co/workspace/My-Workspace~bd24e52e-021c-4145-8b09-90ac08d0be89/collection/22817958-51d7296c-7a00-471c-8d6a-334e16ea70f0?action=share&source=copy-link&creator=22817958
```

### Health Check

```http
GET /healthcheck
```

**Response:**
```json
{
  "message": "OK!!"
}
```

### Authentication Endpoints

#### Register User

```http
POST /api/v1/user/register
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "password": "securepassword"
}
```

#### Login User

```http
POST /api/v1/user/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "securepassword"
}
```

### Book Management Endpoints

> **Note:** All book endpoints require authentication(except book list endpoints). Include the JWT token in the Authorization header:
> ```
> Authorization: Bearer <your-jwt-token>
> ```

#### List Book

```http
GET /api/v1/books?page=1&limit=1&search=programming
Content-Type: application/json
```

#### Create Book

```http
POST /api/v1/books
Content-Type: application/json
Authorization: Bearer <token>

{
  "title": "The Go Programming Language",
  "author": "Alan Donovan",
  "isbn": "978-0134190440",
  "category": "Programming",
  "quantity": 5
}
```

#### Update Book

```http
PUT /api/v1/books/update/{book-id}
Content-Type: application/json
Authorization: Bearer <token>

{
  "title": "Updated Book Title",
  "author": "Updated Author",
  "isbn": "978-0134190440",
  "category": "Updated Category",
  "quantity": 10
}
```

#### Delete Book

```http
DELETE /api/v1/books/delete/{book-id}
Authorization: Bearer <token>
```

### Lending Book Management Endpoints

> **Note:** All Lending book endpoints require authentication. Include the JWT token in the Authorization header:
> ```
> Authorization: Bearer <your-jwt-token>
> ```

#### Borrow Book

```http
POST /api/v1/books/{book-id}/borrow
Content-Type: application/json
Authorization: Bearer <token>
```

#### Return Book

```http
POST /api/v1/books/{lending-id}/return
Content-Type: application/json
Authorization: Bearer <token>
```

## 🗄️ Database Schema

### Users Table

| Column     | Type      | Description          |
|------------|-----------|----------------------|
| id         | VARCHAR   | Primary key (UUID)   |
| name       | VARCHAR   | User's full name     |
| email      | VARCHAR   | User's email address |
| password   | VARCHAR   | Hashed password      |
| role       | VARCHAR   | User's role          |
| created_at | TIMESTAMP | Creation timestamp   |
| updated_at | TIMESTAMP | Last update time     |

### Books Table

| Column     | Type      | Description        |
|------------|-----------|--------------------|
| id         | VARCHAR   | Primary key (UUID) |
| title      | VARCHAR   | Book title         |
| author     | VARCHAR   | Book author        |
| isbn       | VARCHAR   | ISBN number        |
| category   | VARCHAR   | Book category      |
| quantity   | INTEGER   | Available quantity |
| created_at | TIMESTAMP | Creation timestamp |
| created_by | VARCHAR   | Creator user name  |
| updated_at | TIMESTAMP | Last update time   |
| updated_by | VARCHAR   | Last updater name  |
| deleted_at | TIMESTAMP | Soft delete time   |
| deleted_by | VARCHAR   | Deleter user name  |

### Lending Records Table

| Column      | Type      | Description        |
|-------------|-----------|--------------------|
| id          | VARCHAR   | Primary key (UUID) |
| user_id     | VARCHAR   | User lending       |
| book_id     | VARCHAR   | Book Lending       |
| borrow_date | VARCHAR   | Borrow timestamp   |
| return_date | VARCHAR   | Return timestamp   |
| status      | VARCHAR   | Lending status     |
| created_at  | TIMESTAMP | Creation timestamp |
| updated_at  | TIMESTAMP | Last update time   |

## 🔧 Configuration

The application uses Viper for configuration management. You can configure the application using:

1. Environment variables
2. Configuration files (JSON, YAML, TOML)
3. Remote configuration systems

Key configuration options:

- `APP_NAME`: Application name
- `APP_ENV`: Environment (local, development, staging, production)
- `PORT`: Server port
- `DB_*`: Database connection parameters
- `JWT_KEY`: JWT signing secret
- `CONFIG_ID`: Configuration identifier

## 🧪 Testing

Run the tests:

```bash
go test ./...
```

Run tests with coverage:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 📝 Logging

The application includes comprehensive logging with different log levels:

- `DEBUG`: Detailed information for debugging
- `INFO`: General information messages
- `ERROR`: Error conditions

Logs include request IDs for tracing and monitoring.

## 🚦 Middleware

The application includes several middleware components:

- **CORS**: Cross-Origin Resource Sharing support
- **Error Handler**: Centralized error handling and recovery
- **Context ID**: Request tracing with unique identifiers
- **Authentication**: JWT token validation
- **Authorization**: Role-based access control
