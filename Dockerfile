# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.23.4 AS build
ARG APP
WORKDIR /src

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build-${APP}

FROM alpine:latest AS final
ARG APP
WORKDIR /spotify

COPY --from=build /src/.build/spotify.${APP} .
COPY --from=build /src/config.toml .

ENV ENTRYPOINT_CMD=/spotify.${APP}

ENTRYPOINT [ "sh", "-c", "$ENTRYPOINT_CMD" ]
