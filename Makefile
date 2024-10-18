INP_DIR = ./proto/v1/
OUT_DIR = ./internal/pb

.PHONY: proto
proto:
	@protoc -I=$(INP_DIR) \
		--go_out=paths=source_relative:$(OUT_DIR) \
		--go-grpc_out=paths=source_relative:$(OUT_DIR) \
		$(INP_DIR)/*.proto

.PHONY: docker
docker:
	docker compose down --remove-orphans && \
	docker rmi ml/go-distributed-cache; \
	docker compose build --force-rm && \
	docker image prune -f && \
	docker compose up -d
