#!/usr/bin/env bash

IMAGE=docker.jw4.us/import:latest
NAME=import
SCRIPTDIR="$(cd "$(dirname "$0")"; pwd -P)"

docker pull ${IMAGE}
docker stop ${NAME}
docker logs ${NAME} &> $(TZ=UTC date +%Y-%m-%d-%H%M-${NAME}.log)
docker rm -v -f ${NAME}

docker run -d \
  --name ${NAME} \
  --restart=always \
  -e IMPORT_LISTEN_ADDRESS=":19980" \
  -e IMPORT_DB_FILE="/etc/import/repo.db" \
  -e IMPORT_PUBLIC_DIR="/public" \
  -e IMPORT_VERBOSE_LOGGING="" \
  -e IMPORT_SAFE_IPS="192.168.199.0/24" \
  -p 19980:19980 \
  -v ${SCRIPTDIR}/config:/etc/import \
  -v ${SCRIPTDIR}/public:/public \
  ${IMAGE}
