#!/bin/bash
rm -rf redis
mkdir redis

if [ ! -d redis/src ]; then
    curl -O http://download.redis.io/redis-stable.tar.gz
    tar xvzf redis-stable.tar.gz -C .
    rm redis-stable.tar.gz
fi

ls -a
cd redis-stable
make install
