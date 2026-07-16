.PHONY: dev test build lint migrate-up migrate-status learning-content-validate learning-content-verify check-compose-exposure clean

dev:
	docker compose -f docker-compose.yml -f docker-compose.dev.yml up postgres -d
	# 本地混合开发：后台启动 Gateway，前台启动前端 dev server
	# 停止时运行 `make clean`
	DATABASE_URL=$${LOCAL_DATABASE_URL:-postgres://user:pass@localhost:5432/gogopher?sslmode=disable} go run ./cmd/migrate up
	go run ./cmd/sandbox &
	DATABASE_URL=$${LOCAL_DATABASE_URL:-postgres://user:pass@localhost:5432/gogopher?sslmode=disable} LEARNING_SLICE_ENABLED=true APP_ENV=local go run ./cmd/gateway &
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

learning-content-validate:
	go run ./cmd/learning-content validate --activity-set m1-first-slice

learning-content-verify:
	go run ./cmd/learning-content verify --release-dir content/learning/releases/m1-first-slice-v4 --web-dist web/dist

check-compose-exposure:
	./scripts/check-compose-exposure.sh

clean:
	docker compose down -v
	pkill -f "go run ./cmd/gateway" || true
	pkill -f "go run ./cmd/sandbox" || true
