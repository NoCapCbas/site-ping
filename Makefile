# Development commands

# Start the development environment
dev:
	docker-compose --env-file .env.dev -f docker-compose.dev.yml up --build -d --force-recreate

test:
	docker-compose -f docker-compose.dev.yml exec auth_app python -m pytest

# Start the production environment
prod:
	docker-compose --env-file .env.prod -f docker-compose.prod.yml up --build -d --force-recreate

.PHONY: dev prod

# Display help information
help:
	@echo "Available commands:"
	@echo "  dev    - Start the development environment"
	@echo "  prod   - Start the production environment"
	@echo "  help   - Display this help message"

.PHONY: help
