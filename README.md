# TaskFlow API

A REST API for project and task management built with Go, following Clean Architecture principles.

## Tech Stack

- **Language:** Go 1.22+
- **Router:** Chi v5
- **Database:** PostgreSQL (pgx + SQLC)
- **Cache:** Redis
- **Auth:** JWT + bcrypt
- **Migrations:** golang-migrate
- **Logging:** zerolog
- **Containerization:** Docker + Docker Compose

## Architecture

The project follows Clean Architecture with clear separation of concerns:

```
taskflow/
├── cmd/api/          # Entry point
├── config/           # Environment configuration
├── internal/
│   ├── handler/      # HTTP handlers (request/response)
│   ├── service/      # Business logic
│   ├── repository/   # Data access layer (SQLC)
│   ├── middleware/   # JWT auth, logging
│   └── model/        # Domain models
├── db/               # SQLC generated code
├── migrations/       # SQL migrations
├── queries/          # SQL queries for SQLC
└── docker-compose.yml
```

**Request flow:**
```
HTTP Request → Middleware → Handler → Service → Repository → PostgreSQL
```

## API Endpoints

### Auth (public)
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/register` | Register a new user |
| POST | `/auth/login` | Login and get JWT token |

### Users (protected)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/users/me` | Get current user profile |
| PUT | `/api/users/me` | Update current user profile |

### Projects (protected)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/projects` | List user's projects |
| POST | `/api/projects` | Create a new project |
| GET | `/api/projects/{id}` | Get project by ID |
| PUT | `/api/projects/{id}` | Update project |
| DELETE | `/api/projects/{id}` | Delete project |

### Project Members (protected)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/projects/{id}/members` | List project members |
| POST | `/api/projects/{id}/members` | Add member to project |
| DELETE | `/api/projects/{id}/members/{userId}` | Remove member from project |

### Tasks (protected)
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/projects/{id}/tasks` | List project tasks |
| POST | `/api/projects/{id}/tasks` | Create a task |
| GET | `/api/projects/{id}/tasks/{tid}` | Get task by ID |
| PUT | `/api/projects/{id}/tasks/{tid}` | Update task |
| DELETE | `/api/projects/{id}/tasks/{tid}` | Delete task |

## Getting Started

### Prerequisites

- Go 1.22+
- Docker & Docker Compose
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [sqlc](https://sqlc.dev)

### Installation

**1. Clone the repository**
```bash
git clone https://github.com/kevstaa/taskflow.git
cd taskflow
```

**2. Set up environment variables**
```bash
cp .env.example .env
# Edit .env with your values
```

**3. Start the infrastructure**
```bash
make docker-up
```

**4. Run database migrations**
```bash
make migrate-up
```

**5. Start the server**
```bash
make run
```

The API will be available at `http://localhost:8080`.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `JWT_SECRET` | Secret key for JWT signing | - |
| `DATABASE_URL` | PostgreSQL connection string | - |
| `REDIS_URL` | Redis connection address | - |
| `ENVIRONMENT` | Application environment | `development` |

## Usage Examples

**Register a user**
```bash
curl -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username": "example", "email": "example@example.com", "password": "password123"}'
```

**Login**
```bash
curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "example@example.com", "password": "password123"}'
```

**Create a project** (requires JWT)
```bash
curl -X POST http://localhost:8080/api/projects \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{"name": "My Project", "description": "Project description"}'
```

**Create a task** (requires JWT)
```bash
curl -X POST http://localhost:8080/api/projects/<project-id>/tasks \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{"title": "My Task", "description": "Task description", "status": "todo"}'
```

## Available Make Commands

```bash
make run          # Start the server
make test         # Run tests
make docker-up    # Start PostgreSQL and Redis containers
make docker-down  # Stop containers
make migrate-up   # Apply pending migrations
make migrate-down # Rollback last migration
make sqlc         # Regenerate SQLC code
```

## Database Schema

```
users
├── id (UUID, PK)
├── username (VARCHAR, unique)
├── email (VARCHAR, unique)
├── password_hash (VARCHAR)
├── created_at (TIMESTAMP)
└── updated_at (TIMESTAMP)

projects
├── id (UUID, PK)
├── name (VARCHAR)
├── description (TEXT)
├── owner_id (UUID, FK → users)
├── created_at (TIMESTAMP)
└── updated_at (TIMESTAMP)

project_members
├── project_id (UUID, FK → projects)
├── user_id (UUID, FK → users)
├── role (VARCHAR)
└── joined_at (TIMESTAMP)

tasks
├── id (UUID, PK)
├── title (VARCHAR)
├── description (TEXT)
├── status (VARCHAR: todo | in_progress | done)
├── project_id (UUID, FK → projects)
├── assignee_id (UUID, FK → users, nullable)
├── created_at (TIMESTAMP)
└── updated_at (TIMESTAMP)
```

## License

MIT
