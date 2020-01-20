.PHONY: connect
connect: producer consumer

.PHONY: up
up:
	docker-compose down || true
	docker-compose up
.PHONY: down
down:
	docker-compose down

.PHONY: producer
producer:
	docker-compose exec primary1 nfdc face create udp://172.22.1.21
	docker-compose exec primary1 nfdc route add /producer/1/a udp://172.22.1.21
.PHONY: consumer
consumer:
	docker-compose exec consumer1a nfdc face create udp://172.22.1.1
	docker-compose exec consumer1a nfdc route add / udp://172.22.1.1
