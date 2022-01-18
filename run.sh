#!/bin/bash

# Start dalc-test
docker run --platform linux/amd64 --network=docker_localnet jbowen93/dalc:test2 dalc start
# docker run -it --platform linux/amd64 --network=docker_localnet jbowen93/dalc:test bash
