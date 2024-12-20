# syntax=docker/dockerfile:1
FROM --platform=$BUILDPLATFORM golang:1.23.4 AS build
ARG APP
WORKDIR /src

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 make build-${APP}

FROM alpine:latest AS final
ARG APP
USER root

COPY --from=build /src/.build/spotify.${APP} /bin/
COPY --from=build /src/.build/config.toml /bin/

ENV ENTRYPOINT_CMD=/bin/spotify.${APP}

ENTRYPOINT [ "sh", "-c", "$ENTRYPOINT_CMD" ]
