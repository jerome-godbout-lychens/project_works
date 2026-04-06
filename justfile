
docker-run:
    docker compose -f ./docker-compose.yaml up -d

docker-build:
    docker compose -f ./docker-compose.yaml build

docker-build-run: docker-build docker-run

database:
    docker compose -f ./docker-compose.yaml up -d database file_storage

docker-stop:
    docker compose -f ./docker-compose.yaml down

docker-stop-remove:
    docker compose -f ./docker-compose.yaml down --volumes --remove-orphans

docker-logs:
    docker compose -f ./docker-compose.yaml logs -f backend

shell_backend:
    docker compose -f ./docker-compose.yaml exec backend /bin/sh

clear-local-storage: docker-stop
    rm -fr local_storage_data/postgres/*
    rm -fr local_storage_data/seaweedfs/*
    git restore local_storage_data/postgres/
    git restore local_storage_data/seaweedfs/
