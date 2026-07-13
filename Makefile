.PHONY: dev test build lint migrate-up migrate-status clean

dev:
	docker compose up postgres redis -d
	# 本地混合开发：后台启动 Go 服务，前台启动前端 dev server
	# 停止时运行 `make clean`
	go run ./src/services/sandbox-engine/main.go &
	go run ./src/services/gateway/main.go &
	cd web && npm run dev

test:
	go test ./...

build:
	docker compose build

lint:
	go vet ./...

migrate-up:
	go run ./cmd/migrate up

migrate-status:
	go run ./cmd/migrate status

clean:
	docker compose down -v
	pkill -f "go run ./src/services" || true
