.PHONY: build dev run down logs restart

build:
	docker-compose build

db:
	docker-compose up -d db

dev:
	docker-compose --profile dev up

run:
	docker-compose --profile prod up -d

down:
	docker-compose down

logs:
	docker-compose logs -f

restart: down build dev