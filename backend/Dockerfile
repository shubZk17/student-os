FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
 && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

# Alpine (not distroless) so platforms that run overrides via /bin/sh -c work.
FROM alpine:3.20
RUN apk add --no-cache ca-certificates
COPY --from=build /out/ /app/
USER nobody
ENV ENV=production PORT=8080
EXPOSE 8080
# Override with /app/worker for the ingestion cron job.
CMD ["/app/server"]
