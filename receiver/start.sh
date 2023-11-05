#docker network create testudp
docker build -t receiver . && docker run --network=testudp --rm --name receiver receiver

