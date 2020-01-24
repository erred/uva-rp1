.PHONY: connect
connect: producer consumer

.PHONY: up
up:
	docker-compose down || true
	docker-compose up
.PHONY: down
down:
	docker-compose down

.PHONY: producer producer1a producer1b producer2a producer3a producer4a
producer: producer1a producer2a producer3a producer4a
producer1a:
	docker-compose exec primary1 nfdc face create udp://172.26.1.21
	docker-compose exec primary1 nfdc route add /producer/1/a udp://172.26.1.21
producer1b:
	docker-compose exec primary1 nfdc face create udp://172.26.1.22
	docker-compose exec primary1 nfdc route add /fileserve udp://172.26.1.22
producer2a:
	docker-compose exec primary2 nfdc face create udp://172.26.2.21
	docker-compose exec primary2 nfdc route add /producer/2/a udp://172.26.2.21
producer3a:
	docker-compose exec primary3 nfdc face create udp://172.26.3.21
	docker-compose exec primary3 nfdc route add /producer/3/a udp://172.26.3.21
producer4a:
	docker-compose exec primary4 nfdc face create udp://172.26.4.21
	docker-compose exec primary4 nfdc route add /producer/4/a udp://172.26.4.21
.PHONY: consumer consumer1a consumer2a consumer3a consumer4a
consumer: consumer1a consumer2a consumer3a consumer4a
consumer1a:
	docker-compose exec consumer1a nfdc face create udp://172.26.1.1
	docker-compose exec consumer1a nfdc route add / udp://172.26.1.1
consumer2a:
	docker-compose exec consumer2a nfdc face create udp://172.26.2.1
	docker-compose exec consumer2a nfdc route add / udp://172.26.2.1
consumer3a:
	docker-compose exec consumer3a nfdc face create udp://172.26.3.1
	docker-compose exec consumer3a nfdc route add / udp://172.26.3.1
consumer4a:
	docker-compose exec consumer4a nfdc face create udp://172.26.4.1
	docker-compose exec consumer4a nfdc route add / udp://172.26.4.1
