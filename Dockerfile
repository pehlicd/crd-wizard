# syntax=docker/dockerfile:1
FROM scratch
ENTRYPOINT ["/crd-wizard"]
COPY crd-wizard /
