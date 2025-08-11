FROM golang:1.20-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o log-server .

FROM alpine:3.18
WORKDIR /app
COPY --from=build /app/log-server /app/log-server
VOLUME /data
EXPOSE 8081
CMD ["/app/log-server"]
