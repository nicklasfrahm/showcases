FROM golang AS build
ARG VERSION
ARG SERVICE
WORKDIR /app
COPY go.* ./
RUN go mod download
ADD Makefile /app
ADD pkg /app/pkg
ADD cmd/$SERVICE /app/cmd/$SERVICE
RUN VERSION=$VERSION make bin/$SERVICE \
    && mv bin/$SERVICE app

FROM gcr.io/distroless/base:nonroot AS run
COPY --from=build /app/app /
ENTRYPOINT [ "/app" ]
