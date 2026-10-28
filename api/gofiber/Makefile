project_name = duckz
image_name = duckz:latest
localdev = -f docker-compose.yml -f docker-compose-localdev.yml

check-local:
	docker compose $(localdev) config --quiet

run-local:
	go run main.go

requirements:
	go mod tidy

mod-upgrade:
	go get -u ./...
	go mod tidy
	go mod download
	go mod vendor

mod:
	go mod tidy
	go mod download
	go mod vendor

clean-packages:
	go clean -modcache

clean-all:
	 go clean -i -n -r -x -cache -modcache

test-build:
	go build
	go clean

test:
	go test -count=1 ./...

# needs the local stack running (make start)
test-smoke:
	go test -count=1 -v -run TestLiveSmoke ./router/

up:
	make up-silent
	make shell

build:
	docker build -t $(image_name) .

build-no-cache:
	docker build --no-cache -t $(image_name) .

up-silent:
	make delete-container-if-exist
	docker run -d -p 3000:3000 --name $(project_name) $(image_name) ./app

up-silent-prefork:
	make delete-container-if-exist
	docker run -d -p 3000:3000 --name $(project_name) $(image_name) ./app -prod

delete-container-if-exist:
	docker stop $(project_name) || true && docker rm $(project_name) || true

shell:
	docker exec -it $(project_name) /bin/sh

stop:
	docker compose down -v
	
start:
	docker compose up -d

stop-admin:
	docker compose $(localdev) down -v admin

start-admin:
	docker compose $(localdev) up -d admin

stop-db:
	docker compose down -v db adminer

start-db:
	docker compose up -d db adminer

stop-gofiber:
	docker compose $(localdev) down -v db adminer cache cachegui gofiber

start-gofiber:
	docker compose $(localdev) up -d db adminer cache cachegui gofiber

stop-gofiber2:
	docker compose $(localdev) down -v gofiber2

start-gofiber2:
	docker compose $(localdev) up --build -d gofiber2

start-watch:
	docker compose $(localdev) watch db adminer cache cachegui gofiber

start-es:
	docker compose $(localdev) up -d elasticsearch kibana

restart:
	docker compose restart gofiber
