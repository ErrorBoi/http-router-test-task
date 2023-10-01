#!/bin/bash

# Build services

mkdir -p bin

echo -e "Build client ... "
go build -o bin/client client/main.go
echo "done"

echo -e "Build proxy ... "
go build -o bin/proxy ./proxy
echo "done"

echo -e "Build recipient ... "
go build -o bin/recipient recipient/main.go
echo "done"

# Start services

mkdir -p log
rm -f log/*.log

./bin/recipient -i 1 -p 8051 &> ./log/recip1.log & RECIP1_PID=$!
./bin/recipient -i 2 -p 8052 &> ./log/recip2.log & RECIP2_PID=$!
./bin/recipient -i 3 -p 8053 &> ./log/recip3.log & RECIP3_PID=$!
echo "Started recipients"

./bin/proxy -p 8050 -r 8051,8052,8053 &> ./log/proxy.log & PROXY_PID=$!
echo "Started proxy"

./bin/client -p 8050 -l 1000 &> ./log/client.log
echo "Finish session"

# Stop all

pkill proxy
echo "Stopped proxy"

pkill recipient
echo "Stopped recipients"

echo "All services stopped"
