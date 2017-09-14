FROM ubuntu:16.04
COPY l2flat-helper /usr/bin/
RUN apt-get update && \
    apt-get install --no-install-recommends -y \
    arping && \
    rm -rf /var/lib/apt/lists/*
CMD ["l2flat-helper"]
