#!/bin/bash

# Build dalc-test
docker build --platform linux/amd64 -f docker/Dockerfile -t jbowen93/dalc:test .