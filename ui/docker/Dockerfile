FROM ubuntu:bionic AS base

RUN apt-get update -y && \
    apt-get install -y \
        build-essential \
        curl \
        git

FROM base AS repo
WORKDIR /repo

ARG STATIC_DIRECTORY
ARG BASE_PATH
ARG API_BASE_PATH

ENV PATH="/node/bin:${PATH}"
RUN mkdir /node && \
    curl -sL https://git.io/n-install | N_PREFIX=/node bash -s -- -q && \
    npm i -g yarn

COPY . /repo/influxdb/
WORKDIR /repo/influxdb/ui

# these are all run together as docker's caching mechanism
# makes big steps like yarn install expensive
RUN yarn install --production=false && \
    yarn generate && \
    INFLUXDB_SHA=remote $(npm bin)/webpack --config ./webpack.prod.ts --bail && \
    rm -rf ./node_modules

FROM nginx:stable-alpine
COPY --from=repo /repo/influxdb/ui/build /usr/share/nginx/html
COPY ui/docker/nginx.conf /etc/nginx/nginx.conf
