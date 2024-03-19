FROM debian

COPY ./e2e-test /usr/local/bin

RUN \
    apt-get update \
 && apt-get install -y --no-install-recommends \
    tini \
    ca-certificates \
    curl \
    bash \
 && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/usr/bin/tini", "--", "e2e-test"]
