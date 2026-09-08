# syntax=docker/dockerfile:1.7
# Multi-stage Linux OCI image for CloudOps API + worker.
# Final stages have no compiler, no package manager, and no secrets.

FROM --platform=linux/amd64 public.ecr.aws/docker/library/golang:1.27-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api
RUN go build -trimpath -ldflags="-s -w" -o /out/worker ./cmd/worker

FROM --platform=linux/amd64 public.ecr.aws/lambda/provided:al2023 AS lambda
COPY --from=build --chown=1000:1000 /out/api ${LAMBDA_TASK_ROOT}/api
COPY --from=build --chown=1000:1000 /out/worker ${LAMBDA_TASK_ROOT}/worker
# Non-root image user. Lambda may remap the runtime user.
USER 1000:1000
WORKDIR ${LAMBDA_TASK_ROOT}
CMD [ "api" ]

# Local / CI HTTP image used to prove PID 1 + SIGTERM behaviour.
FROM --platform=linux/amd64 gcr.io/distroless/static-debian12:nonroot AS http
WORKDIR /app
COPY --from=build --chown=65532:65532 /out/api /app/cloudops-api
USER 65532:65532
ENV HOME=/tmp
EXPOSE 8080
ENTRYPOINT ["/app/cloudops-api"]
