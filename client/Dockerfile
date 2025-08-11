FROM golang:1.20-alpine AS build
WORKDIR /app
COPY . .
RUN go build -o client .

FROM alpine:3.18
WORKDIR /app
COPY --from=build /app/client /app/client
CMD ["/app/client"]
