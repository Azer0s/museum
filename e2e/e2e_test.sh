#!/usr/bin/env bash

echo "Running e2e test"

echo "Building museum"
go build ../cmd/museum
echo "Building museum successful ✅"

echo "Starting etcd"
docker compose up -d
echo "Starting etcd successful ✅"

echo "Starting museum server"
export ETCD_HOST=localhost:2379
./museum server &
echo "Starting museum server successful ✅"

# wait for server to start
sleep 5

cp ../examples/nginx.exhibit ./nginx.exhibit
sed -i 's/2h/20s/g' ./nginx.exhibit

echo "Creating exhibit"
out=$(./museum create ./nginx.exhibit | grep 'http' | awk -F' ' '{print $2}')
curl -s "$out" > /dev/null
echo "Creating exhibit successful ✅"

# wait for exhibit to be created
sleep 5

# curl to check if result contains "nginx" (which means exhibit is running)
if curl -s "$out" | grep -q "nginx"; then
    echo "Exhibit is running ✅"
else
    echo "Exhibit is not running"
    exit 1
fi

# check in docker if exhibit is running
if [ "$(docker ps -q -f name=my-site_nginx)" ]; then
    echo "Docker container is running ✅"
else
    echo "Docker container is not running"
    exit 1
fi

# wait for exhibit to be deleted
sleep 30

# check in docker if exhibit is deleted
if [ "$(docker ps -q -f name=my-site_nginx)" ]; then
    echo "Docker container is running"
    exit 1
else
    echo "Docker container is not running anymore ✅"
fi

fuser -k 8080/tcp

rm ./nginx.exhibit
rm ./museum

docker compose down
docker rm test-etcd-1