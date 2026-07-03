FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o household_calculator .


FROM alpine:latest
RUN apk add --no-cache libc6-compat

WORKDIR /app

COPY --from=builder /app/household_calculator .

VOLUME ["/app/data"]
WORKDIR /app/data

ENTRYPOINT ["/app/household_calculator"]
CMD ["list"]
