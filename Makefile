start_otel:
	@docker

# e2e 测试

orm_e2e:
	docker compose down
	docker compose up -d
	go test -race ./orm/... -tags=e2e
	docker compose down