
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

shell-backend:
    docker compose -f ./docker-compose.yaml exec backend /bin/sh

integration-test-build:
    docker build -t project-works-integration-test:latest -f Dockerfile_tests .

integration-test-shell:
    docker run -it --network project_works_app_network project-works-integration-test:latest

integration-test-run:
    docker run --rm --network project_works_app_network project-works-integration-test:latest -c "/app/project-works-integration-test -base-url http://project_works_backend:8088 -super-admin-api-key super-admin-dev-key-do-not-use-in-production"

clear-local-storage: docker-stop
    rm -fr local_storage_data/postgres/*
    rm -fr local_storage_data/seaweedfs/*
    git restore local_storage_data/postgres/
    git restore local_storage_data/seaweedfs/
