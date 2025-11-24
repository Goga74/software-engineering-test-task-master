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
- **[API_KEY_AUTH.md](API_KEY_AUTH.md)** - API Key authentication implementation guide
- **[K8S_DEPLOYMENT.md](K8S_DEPLOYMENT.md)** - Kubernetes deployment guide with manifests - branch feature/kubernetes-manifests

### Implementation Notes

**Author**: Igor Zamiatin

All tasks are solved, except 2 tasks from "Bonus Points" section:
- Terraform
- CD pipeline created, but has some errors.
I am in the process of solving CD pipeline.

### Branch Strategy

**Note on branching approach**: While creating separate feature branches for each individual task (Task1, Task2, etc.) would be a best practice, this approach was deemed excessive for the initial simple tasks. Instead, task references were clearly indicated in commit messages (e.g., "Task1: ...", "Task2: ...") to maintain traceability.

The `bonus_points` branch serves as an aggregation point for all bonus task implementations.

### Code Quality Improvements

GitHub Actions CI pipeline was implemented early in the development process. Following its integration, comprehensive code refactoring was performed to address linting and formatting issues. For example, see the error handling improvements in the repository layer: [users.go#L34](https://github.com/Goga74/software-engineering-test-task-master/blob/main/internal/repository/users.go#L34)

### Scope Limitations

Tasks requiring deployment to cloud services (which necessitate corresponding cloud accounts) - such as Terraform infrastructure provisioning and CD pipeline to production - have not been implemented at this time. However, all necessary groundwork including:
- Kubernetes manifests
- Configuration 
management
- Docker optimization
- Comprehensive testing

...has been completed to facilitate future cloud deployment when required.

### Documentation

Comprehensive documentation has been created for all implemented features:
- Docker deployment guide
- Testing guide (unit and integration tests)
- Configuration management guide
- Kubernetes deployment guide
- Bonus features documentation (logging, API authentication)

All documentation follows industry best practices and includes troubleshooting guides, examples, and Windows-specific instructions where applicable.

---


