FROM golang:alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .
RUN go build -o backend

FROM busybox:1.36.1-musl

WORKDIR /app

COPY --from=build /app/backend .

CMD [ "./backend" ]