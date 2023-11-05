# docker network create testudp
docker build -t sender . && docker run --network=testudp -v $(pwd):/tmp/ --rm --name sender sender ./myapp test.txt 

