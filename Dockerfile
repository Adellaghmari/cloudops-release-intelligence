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

# provided.al2023 looks for an executable named bootstrap under LAMBDA_TASK_ROOT
# (/var/task/bootstrap). CMD / image_config.command selects api|worker via _HANDLER.
FROM --platform=linux/amd64 public.ecr.aws/lambda/provided:al2023 AS lambda
COPY --from=build /out/api ${LAMBDA_TASK_ROOT}/api
COPY --from=build /out/worker ${LAMBDA_TASK_ROOT}/worker
COPY lambda/bootstrap ${LAMBDA_TASK_ROOT}/bootstrap
RUN chmod 755 ${LAMBDA_TASK_ROOT}/bootstrap ${LAMBDA_TASK_ROOT}/api ${LAMBDA_TASK_ROOT}/worker \
  && test -x ${LAMBDA_TASK_ROOT}/bootstrap \
  && test -x ${LAMBDA_TASK_ROOT}/api \
  && test -x ${LAMBDA_TASK_ROOT}/worker
# Do not set USER here. The AWS Lambda base-image entrypoint must run as the
# image default user; Lambda's execution environment provides the sandbox.
# Non-root posture is proven on the HTTP distroless stage below.
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
