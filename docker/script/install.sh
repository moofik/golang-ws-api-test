#!/usr/bin/env bash

NET_NAME=brainpower
PHP_CONTAINER_NAME=brainpower/php-fpm

. $(dirname $0)/functions.sh
. $(dirname $0)/colored_output.sh

set -e

if [[ -z "$(${SUDO_CMD} docker network inspect ${NET_NAME} > /dev/null 2>&1 && echo 1)" ]]; then
    ${SUDO_CMD} docker network create --subnet 172.85.0.0/24 ${NET_NAME}
    echo -e "${CYAN}Docker network created. Gateway is 172.85.0.1\r\n${NC}"
fi

if [[ ! -f "${DIR}/../../backend/.env" ]]; then
    cp "${DIR}/../../backend/.env.example" "${DIR}/../../backend/.env"
    echo -e "${CYAN}.env file for backend restored\r\n${NC}"
else
    echo -e "${CYAN}.env file already existed! This is good :)"
fi

if [[ ${ENVIRONMENT} -eq "develop" ]]; then
     if [[ ! -f "${DIR}/../docker-compose.yml" ]]; then
        cp "${DIR}/../docker-compose.yml.dist" "${DIR}/../docker-compose.yml"
        echo -e "${CYAN}Docker config created from dist\r\n${NC}"
    fi

    if [[ -z "$(${SUDO_CMD} docker image inspect ${PHP_CONTAINER_NAME} > /dev/null 2>&1 && echo 1)" ]]; then
        bash -c "cd ${DIR}/.. && ${SUDO_CMD} docker-compose build --no-cache"
        echo -e "${CYAN}Docker containers were built\r\n${NC}"
    fi

    # install node modules
    if [[ ! -d "${DIR}/../../backend/node_modules" ]]; then
        bash -c "cd ${DIR}/../.. && MOUNT_DIR=backend docker/yarn install"
        echo -e "${CYAN}Node modules for backend bundle were successfully installed\r\n${NC}"
    fi

    # install backend vendor packages
    if [[ ! -d "${DIR}/../../backend/vendor" ]]; then
        bash -c "cd ${DIR}/../.. && docker/composer install"
        echo -e "${CYAN}Vendor packages for backend bundle were successfully installed\r\n${NC}"
    else
        echo -e "${CYAN}Vendor packages already installed! This is good :)"
    fi

# up containers
    bash -c "cd ${DIR}/../.. && docker/up"
    echo -e "${CYAN}Containers were successfully started\r\n${NC}"

    bash -c "cd ${DIR}/../.. && docker/artisan telescope:install"
    bash -c "cd ${DIR}/../.. && docker/artisan telescope:publish"
    echo -e "${CYAN}Laravel Telescope was successfully installed\r\n${NC}"

    #install ide-helper
    bash -c "cd ${DIR}/../.. && docker/artisan ide-helper:generate"
    bash -c "cd ${DIR}/../.. && docker/artisan ide-helper:models"
    bash -c "cd ${DIR}/../.. && docker/artisan ide-helper:meta"
    echo -e "${CYAN}barryvdh/ide-helper was successfully installed\r\n${NC}"

    # run migrations
    bash -c "cd ${DIR}/../.. && docker/artisan migrate"
    echo -e "${CYAN}Migrations was successfully executed\r\n${NC}"

    ### run fixtures
    bash -c "cd ${DIR}/../.. && docker/artisan db:seed"
    echo -e "${CYAN}Database was successfully seeded\r\n${NC}"
    echo -e "${CYAN}Installation completed\r\n${NC}"
fi