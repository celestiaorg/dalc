#!/bin/bash

set -e 

if [ "$1" = 'celestia' ]; then
    ./celestia "${NODE_TYPE}" --node.store /dalc init

    exec ./"$@" "--"
fi

exec "$@"
