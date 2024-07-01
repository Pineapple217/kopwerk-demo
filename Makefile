docker-build:
	swag fmt
	swag init
	docker build . -t  pineapple217/kopwerk-demo:latest

docker-push:
	docker push pineapple217/kopwerk-demo:latest

docker-update:
	@make --no-print-directory docker-build
	@make --no-print-directory docker-push