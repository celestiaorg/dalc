#!/bin/bash

set -e 

if [ "$1" = 'dalc' ]; then
    ./dalc --home /root init
    exec ./"$@" "--"
fi

exec "$@"
