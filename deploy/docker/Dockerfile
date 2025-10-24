# syntax=docker/dockerfile:1
FROM golang:1.22 AS build
WORKDIR /src
COPY . .
RUN go build -o /out/voltvard ./cmd/voltvard

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /out/voltvard /app/voltvard
COPY configs/site.example.yaml /app/config.yaml
EXPOSE 8080
ENTRYPOINT ["/app/voltvard","-config","/app/config.yaml","-addr",":8080"]
