info:
    just -l

#################
# Backend 

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

docker-backend-logs:
    docker compose -f ./docker-compose.yaml logs -f backend

docker-shell-backend:
    docker compose -f ./docker-compose.yaml exec backend /bin/sh

clear-local-storage: docker-stop
    rm -fr local_storage_data/postgres/*
    rm -fr local_storage_data/seaweedfs/*
    git restore local_storage_data/postgres/
    git restore local_storage_data/seaweedfs/

#################
# Tests

test-build:
    # Build the Docker image for testing. This image includes all necessary dependencies to run unit and integration tests.
    docker build -t project-works-test:latest -f Dockerfile_tests .

###
# Unit Tests
unit-test-shell:
    # Run unit tests in an interactive shell inside the test Docker container. This allows for debugging and inspecting test results directly within the container environment.
    docker run --rm -it -v $(pwd):/workspace -v "$(pwd)/test-results:/app/test-results" -w /app project-works-test:latest

unit-test:
    # Run unit tests assume go compiled code.
    mkdir -p test-results
    gotestsum --jsonfile test-results/unit-tests.json --format testname ./...

unit-test-markdown:
    # Generate a markdown report from the JSON test results.
    ./TestsResultsParser.sh

unit-test-ci: unit-test unit-test-markdown

unit-test-ci-docker:
    # Run unit tests in a Docker container and generate test reports. The test results are stored in the host's test-results directory just like CI pipelines.
    mkdir -p test-results
    docker run --rm -v "$(pwd)/test-results:/app/test-results" -w /app project-works-test:latest -c "just unit-test-ci"

###
# Integrations Tests
integration-test-shell:
    # Run integration tests in an interactive shell inside the test Docker container. This allows for debugging and inspecting test results directly within the container environment.
    mkdir -p test-results
    docker run -it --network project_works_app_network -v "$(pwd)/test-results:/app/test-results" -w /app project-works-test:latest

integration-test-run:
    # Run integration tests against the compiled backend service. This command assumes that the backend service is running and accessible at the specified base URL.
    ./project-works-integration-test -base-url http://project_works_backend:8088 -super-admin-api-key super-admin-dev-key-do-not-use-in-production

integration-test-run-docker:
    # Run integration tests in a Docker container. This command assumes that the backend service is running and accessible at the specified base URL.
    mkdir -p test-results
    docker run --rm --network project_works_app_network -v "$(pwd)/test-results:/app/test-results" -w /app project-works-test:latest -c "just integration-test-run"

#################
# Frontend 

frontend-install:
    cd frontend && npm install && npm run codegen

frontend-run:
    cd frontend && npm run dev
