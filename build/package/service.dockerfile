FROM golang AS build
ARG VERSION
ARG SERVICE
WORKDIR /app
COPY go.* ./
RUN go mod download
ADD Makefile /app
ADD cmd /app/cmd
ADD pkg /app/pkg
RUN VERSION=$VERSION make bin/$SERVICE \
    && mv bin/$SERVICE app

FROM gcr.io/distroless/base:nonroot AS run
COPY --from=build /app/app /
ENTRYPOINT [ "/app" ]
