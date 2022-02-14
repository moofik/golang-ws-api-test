#!/usr/bin/env bash

DIR=$(readlink -e $(dirname $0))
SUDO_CMD=$(test -w /var/run/docker.sock || echo sudo)
PROJECT_DIR="/testtask"
ENVIRONMENT=${ENVIRONMENT:-dev}

mysql() {
    local base_dir=$(dirname ${DIR})
    local work_dir=$(pwd | sed "s:${base_dir}:${PROJECT_DIR}:")

    if [[ ${work_dir} = $(pwd) ]]; then
        work_dir="${PROJECT_DIR}"
    fi

    if [[ ${ENVIRONMENT} -eq "prod" ]]; then
        ${SUDO_CMD} docker run \
            --rm \
            -v ${DIR}/../backend:${PROJECT_DIR} \
            -w ${work_dir} \
            --network testtask \
            testtask/mysql \
            "$@"
    else
        ${SUDO_CMD} docker run \
            -it \
            --rm \
            -v ${DIR}/../backend:${PROJECT_DIR} \
            -w ${work_dir} \
            --network testtask \
            testtask/mysql \
            "$@"
    fi
}
