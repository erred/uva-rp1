
network:
	docker network create --subnet 145.100.111.0/28 test6

watcher:
	docker run --rm -d --network test6 --ip 145.100.111.2 -p 8000:8000 \
		seankhliao/ndn-watcher

primary1:
	docker run --rm -d --network test6 --ip 145.100.111.3 -p 6363:6363 -p 8000:8000 \
		-env WATCHER=145.100.111.2:8000 --name=primary1 seankhliao/ndn-primary
primary2:
	docker run --rm -d --network test6 --ip 145.100.111.4 -p 6363:6363 -p 8000:8000 \
		-env WATCHER=145.100.111.2:8000 --name=primary2 seankhliao/ndn-primary
primary3:
	docker run --rm -d --network test6 --ip 145.100.111.5 -p 6363:6363 -p 8000:8000 \
		-env WATCHER=145.100.111.2:8000 --name=primary3 seankhliao/ndn-primary


producer1:
	docker run --rm -d --network test6 --ip 145.100.111.6 -p 6363:6363 -p 8000:8000 \
		-env NDN_TRAFFIC_SERVER_DISABLED= -v $$(pwd)/producer1:/usr/local/etc/ndn \
		seankhliao/ndn-traffic
	docker exec primary1 nfdc face create udp://145.100.111.6
	docker exec primary1 nfdc route add /producer/1/a udp://145.100.111.6

producer2:
	docker run --rm -d --network test6 --ip 145.100.111.7 -p 6363:6363 -p 8000:8000 \
		seankhliao/ndn-fileserve
	docker exec primary2 nfdc face create udp://145.100.111.7
	docker exec primary2 nfdc route add /fileserve udp://145.100.111.7

secondary3a:
	docker run --rm -d --network test6 --ip 145.100.111.8 \
		-env PRIMARY=145.100.111.5:8000 seankhliao/ndn-secondary

secondary3b:
	docker run --rm -d --network test6 --ip 145.100.111.9 \
		-env PRIMARY=145.100.111.5:8000 seankhliao/ndn-secondary

client3a:
	docker run --rm -d --network test6 --ip 145.100.111.10 \
		-env NDN_TRAFFIC_CLIENT_DISABLED= -env FILESERVE_DISABLED=1 \
		-v $$(pwd)/consumer3a:/usr/local/etc/ndn \
		seankhliao/ndn-fileserve
