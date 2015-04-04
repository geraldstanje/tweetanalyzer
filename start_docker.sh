#!/bin/bash

docker build -t outyet .
boot2docker up && $(boot2docker shellinit) 
boot2docker ip
# docker 
docker run -p 8080:8080 -t outyet
# docker ps
# docker inspect hash
# docker stop hash