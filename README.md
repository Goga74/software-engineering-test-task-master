# Simple CRUD Interface

Rewrite the README according to the application.

The task itself can be found [here](/TASK.md)

## Prerequisites

- [Docker](https://www.docker.com/get-started/)
- [Goose](https://github.com/pressly/goose)
- [Gosec](https://github.com/securego/gosec)

## Getting Started

1. Start database

```
## Via Makefile
make db

## Via Docker
docker-compose up -d db
```

2. Run migrations

```
## Via Makefile
make migrate-up

## Via Goose
DB_DRIVER=postgres
DB_STRING="host=localhost port=5432 user=postgres password=postgres dbname=postgres sslmode=disable"
goose -dir ./migrations $(DB_DRIVER) $(DB_STRING) up
```

3. Run application

```
go run cmd/main.go
```

## Documentation

This project includes comprehensive documentation for various aspects of development, deployment, and testing:

### Task Requirements
- **[TASK.md](TASK.md)** - Original task description and requirements

### Deployment
- **[README_DOCKER.md](README_DOCKER.md)** - Docker deployment guide with optimized multi-stage build process

### Testing
- **[README_UNIT_TESTS.md](README_UNIT_TESTS.md)** - Unit tests guide and how to run them
- **[TESTING.md](TESTING.md)** - Comprehensive testing guide including integration tests with real PostgreSQL database

### Additional Documentation
- **[CONFIG.md](CONFIG.md)** - Database configuration management guide - branch feature/config-refactoring
- **[MIDDLEWARE_LOGGING.md](MIDDLEWARE_LOGGING.md)** - JSON logging middleware documentation - branch bonus_points
- **[API_KEY_AUTH.md](BONUS_API_KEY_AUTH.md)** - API Key authentication implementation guide
- **[K8S_DEPLOYMENT.md](K8S_DEPLOYMENT.md)** - Kubernetes deployment guide with manifests - branch feature/kubernetes-manifests

