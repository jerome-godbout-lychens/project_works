
docker-run:
    docker compose -f ./docker-compose.yaml up -d

docker-build:
    docker compose -f ./docker-compose.yaml build

database:
    docker compose -f ./docker-compose.yaml up -d database file_storage

docker-stop:
    docker compose -f ./docker-compose.yaml down

docker-logs:
    docker compose -f ./docker-compose.yaml logs -f web

clear-local-storage: docker-stop
    rm -fr local_storage_data/postgres/*
    rm -fr local_storage_data/seaweedfs/*
    git restore local_storage_data/postgres/
    git restore local_storage_data/seaweedfs/
