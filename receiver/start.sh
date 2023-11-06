#docker network create testudp
# docker build -t receiver . && docker run --network=testudp --rm --name receiver receiver

docker rm -f receiver
docker build --cache-from receiver -t receiver . && \
# docker run --network=testudp -v $(pwd):/tmp/ --rm --name receiver receiver ./myapp
docker run --rm  --name receiver --network=testudp -v /tmp:/root  receiver myapp
