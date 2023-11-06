# docker network create testudp
docker rm -f sender
docker build -t sender . && \
# docker run --rm  --name sender --network=testudp -v $(pwd):/tmp/  sender ./myapp /tmp/test.txt 
docker run --rm  --name sender --network=testudp -v /Users/alexsh/Downloads:/tmp  sender myapp $1

