#!/bin/bash

set -e 

if [ "$1" = 'celestia' ]; then
    exec ./"$@" "--"
fi

exec "$@"
