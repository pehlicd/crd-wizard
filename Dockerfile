FROM node:24 AS frontend-builder

WORKDIR /app/ui

COPY ui /app/ui

RUN npm install -g npm@latest && \
    npm install --force
RUN npm run build

FROM golang:1.24 AS backend-builder

WORKDIR /app

ARG VERSION=dev
ARG BUILD_DATE=unknown
ARG COMMIT_SHA=unknown

COPY . .

RUN go mod tidy && \
    go mod download

COPY --from=frontend-builder /app/ui/dist /app/ui/dist

RUN rm -rf ./internal/web/static/* && \
    mv ./ui/dist/* ./internal/web/static/

ENV CGO_ENABLED=0

RUN go build \
    -ldflags="-s -w -X github.com/pehlicd/crd-wizard/cmd.versionString=${VERSION} -X github.com/pehlicd/crd-wizard/cmd.buildDate=${BUILD_DATE} -X github.com/pehlicd/crd-wizard/cmd.buildCommit=${COMMIT_SHA}" \
    -o crd-wizard

FROM alpine:3.22.1

COPY --from=backend-builder /app/crd-wizard /usr/local/bin/crd-wizard

ARG PORT=8080

ENTRYPOINT ["crd-wizard"]

CMD ["web", "--port", "${PORT}"]
