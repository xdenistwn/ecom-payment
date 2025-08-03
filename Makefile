create-network:
	docker network create internal_vpc

up:
	docker compose -f docker-compose.yml up -d

stop:
	docker compose -f docker-compose.yml stop

down:
	docker compose -f docker-compose.yml down