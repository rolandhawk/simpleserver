default: build push

build:
	docker build -t rolandhawk/simpleserver .

push:
	docker push rolandhawk/simpleserver

deploy:
	kubectl apply -f deployment.yaml

loadtest:
	kubectl run vegeta-preproduction --rm --attach --restart=Never --image="peterevans/vegeta" -- sh -c \
	"echo 'GET http://10.48.173.170:8080' | vegeta attack -connections 5 -rate=500 -duration=5m | tee results.bin | vegeta report --every 1s"
