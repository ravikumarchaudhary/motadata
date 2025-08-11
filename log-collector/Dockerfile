FROM golang:1.20-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o log-collector .

FROM alpine:3.18
WORKDIR /app
COPY --from=build /app/log-collector /app/log-collector
EXPOSE 6000
CMD ["/app/log-collector"]
