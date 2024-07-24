.phony: redis
redis:
	docker run -d --name redis-stack -p 6379:6379 -p 8001:8001 redis/redis-stack:latest

.phone: redis_web
redis_web:
	open http://localhost:8001/