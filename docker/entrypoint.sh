#!/bin/bash

set -e 

if [ "$1" = 'dalc' ]; then
    exec ./"$@" "--"
fi

exec "$@"
