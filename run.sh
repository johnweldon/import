#!/usr/bin/env bash

IMAGE=docker.jw4.us/import
NAME=import
SCRIPTDIR="$(cd "$(dirname "$0")"; pwd -P)"

docker pull ${IMAGE}
docker stop ${NAME}
docker logs ${NAME} &> $(TZ=UTC date +%Y-%m-%d-%H%M-${NAME}.log)
docker rm -v -f ${NAME}

docker run -d \
  --name ${NAME} \
  --restart=always \
  -p 19980:19980 \
  -v ${SCRIPTDIR}/config:/etc/import \
  ${IMAGE}

