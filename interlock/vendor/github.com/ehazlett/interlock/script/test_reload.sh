#!/bin/bash
ACTION="$1"
COUNT=$2

start() {
    docker pull ehazlett/docker-demo
    
    for i in $(seq 1 $COUNT); do
        docker run --name interlock-bench-$i -P --hostname bench-$i.local ehazlett/docker-demo &
    done
}

remove() {
    docker ps -a | grep docker-demo | grep "interlock-bench-*" | awk '{ print $1;  }' | xargs docker rm -fv 
}

case "$1" in
    start)
        start
        ;;
    remove)
        remove
        ;;
    *)
        echo "Usage: $0 <start|remove> [count]"
        exit 1
esac
