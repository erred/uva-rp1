.PHONY: connect-server
connect-server:
	docker-compose exec primary1 nfdc face create udp://172.20.0.12
	docker-compose exec primary1 nfdc route add /rp1/server1 udp://172.20.0.12
	docker-compose exec primary2 nfdc face create udp://172.20.0.22
	docker-compose exec primary2 nfdc route add /rp1/server2 udp://172.20.0.22
	docker-compose exec primary3 nfdc face create udp://172.20.0.32
	docker-compose exec primary3 nfdc route add /rp1/server3 udp://172.20.0.32
