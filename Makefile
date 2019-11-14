NAME = todo
ENTRYPOINT = "cmd/${NAME}/main.go"
BUILD_DIR = build
TMP_DIR = ${BUILD_DIR}/tmp
DSN ?= postgres://postgres:test@localhost:15432/todo?sslmode=disable

default: build

.PHONE: help
help: Makefile
	@echo
	@echo " Choose a command to run in "$(NAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/ /'
	@echo

.PHONY: install build test
## install: Install missing dependencies
install: go-mod-download
## build: Build development version
build: install go-compile-dev
## build-release: Build release verion of the binary
build-release: install go-compile-release
## test: Run tests
test: install go-test

.PHONY: go-mod-download go-compile-dev go-compile-release go-test go-vet go-fmt-chk

go-mod-download:
	@echo " > Install dependencies if necessary"
	@go mod download

go-compile-dev:
	@echo " > Compiling development version..."
	@go build -o "${BUILD_DIR}/${NAME}" $(ENTRYPOINT)

go-compile-release:
	@echo " > Compiling release version..."
	@CGO_ENABLED=0 GOOS=linux \
		go build -a -installsuffix cgo -ldflags '-extldflags "-static"' -o "${BUILD_DIR}/${NAME}" $(ENTRYPOINT)

go-test:
	go test -v -count=1 --race ./...

go-vet: $(TMP_DIR)
	rm -rf ${TMP_DIR}/go_vet.txt
	@go vet ./... 2>> ${TMP_DIR}/go_vet.txt; \
	if [ -s ${TMP_DIR}/go_vet.txt ]; then \
		echo "\033[31m[go vet] failed, please look at ${TMP_DIR}/go_vet.txt for details:\033[0m"; \
	cat ${TMP_DIR}/go_vet.txt; \
	exit 1; \
	else \
		echo "\033[32m[go vet] Everything is fine\033[0m"; \
	fi

go-fmt-chk:
	@ if [ -n "$$(gofmt -l .)" ]; then \
		echo "\033[31m[gofmt] some files need to be formatted:\n$$(gofmt -l .)\033[0m"; \
		exit 1; \
	else \
		echo "\033[32m[gofmt] Formatting is fine\033[0m"; \
	fi

## db-migrate: migrate database to the latest version
db-migrate:
	docker run --rm -v ${PWD}/migrations:/migrations \
		--network host migrate/migrate:v4.7.0 -path=/migrations/ -database ${DSN} \
		up

## db-create-migration: create new migration
db-create-migration:
	@read -p "Enter new migration name: " migrationName; \
	docker run --rm -v ${PWD}/migrations:/migrations \
		-u $$(id -u ${USER}):$$(id -g ${USER}) \
		--network host migrate/migrate:v4.7.0 \
		-path=/migrations/ -database ${DSN} \
		create -ext sql \
		-dir /migrations $$migrationName
