#!/usr/bin/env bash

IMAGE=docker.jw4.us/import:latest
NAME=import
PORT=19980
SCRIPTDIR="$(cd "$(dirname "$0")"; pwd -P)"

docker pull ${IMAGE}
docker stop ${NAME}
docker logs ${NAME} &> $(TZ=UTC date +%Y-%m-%d-%H%M-${NAME}.log)
docker rm -v -f ${NAME}

docker run -d \
  --name ${NAME} \
  --restart=always \
  -e IMPORT_DB_FILE="/etc/import/repo.db" \
  -e IMPORT_SAFE_IPS="192.168.199.0/24" \
  -e IMPORT_VERBOSE_LOGGING="" \
  -e PORT="${PORT}" \
  -p ${PORT}:${PORT} \
  -v ${SCRIPTDIR}/config:/etc/import \
  ${IMAGE}
