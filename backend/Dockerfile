# Copyright 2021-2024 The Kubeflow Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# Build arguments
ARG SOURCE_CODE=.
ARG TARGETOS=linux
ARG TARGETARCH=amd64

FROM registry.access.redhat.com/ubi9/go-toolset:1.26.3-1777889793@sha256:1c1259373e6feb4b57de490452379c40888cf6e876154cd2ace17eae9c64a7ea AS builder

ARG TARGETOS
ARG TARGETARCH

USER root

RUN dnf install -y clang

COPY ${SOURCE_CODE}/go.mod ./
COPY ${SOURCE_CODE}/go.sum ./
COPY ${SOURCE_CODE}/kubernetes_platform/go.mod ./kubernetes_platform/go.mod
COPY ${SOURCE_CODE}/api/go.mod ./api/go.mod

RUN GO111MODULE=on go mod download

# Copy the source
COPY ${SOURCE_CODE}/ ./

# Set CC per target arch for CGO cross-compilation (required by mattn/go-sqlite3).
# clang targets any architecture via --target, avoiding the need for QEMU emulation.
#
# TODO: switch GOFIPS140 from "latest" to "certified" once go-toolset ships a certified.txt snapshot.
# As of go-toolset 1.26.2-1.el9_8, only v1.0.0 and v1.26.0 snapshots exist; no certified.txt.
# Verify with: docker run --rm <go-toolset-image> ls /usr/lib/golang/lib/fips140/certified.txt
RUN if [ "${TARGETARCH}" = "amd64" ]; then CC="clang --target=x86_64-redhat-linux-gnu"; \
    elif [ "${TARGETARCH}" = "arm64" ]; then CC="clang --target=aarch64-redhat-linux-gnu"; fi && \
    GO111MODULE=on CGO_ENABLED=1 CC="${CC}" GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
      GOFIPS140=latest \
      go build -o /bin/apiserver ./backend/src/apiserver/ && \
    dnf clean all

# 2. Compile preloaded pipeline samples
FROM registry.access.redhat.com/ubi9/python-311 AS compiler

ARG TARGETOS TARGETARCH

USER root
RUN dnf install -y python3-setuptools jq
RUN wget https://bootstrap.pypa.io/get-pip.py && python3 get-pip.py
COPY backend/requirements.txt .
RUN python3 -m pip install -r requirements.txt --no-cache-dir

# Downloading Argo CLI so that the samples are validated
ENV ARGO_VERSION=v3.6.7
RUN curl -sLO https://github.com/argoproj/argo-workflows/releases/download/${ARGO_VERSION}/argo-${TARGETOS:-linux}-${TARGETARCH:-amd64}.gz && \
    gunzip argo-${TARGETOS:-linux}-${TARGETARCH:-amd64}.gz && \
    chmod +x argo-${TARGETOS:-linux}-${TARGETARCH:-amd64} && \
    mv ./argo-${TARGETOS:-linux}-${TARGETARCH:-amd64} /usr/local/bin/argo

WORKDIR /
COPY ./samples /samples
COPY backend/src/apiserver/config/sample_config.json /samples/

# Compiling the preloaded samples.
# The default image is replaced with the GCR-hosted python image.
RUN set -e; \
    < /samples/sample_config.json jq ".pipelines[].file" --raw-output | while read pipeline_yaml; do \
    pipeline_py="${pipeline_yaml%.yaml}"; \
    echo "Compiling: \"$pipeline_py\"" && python3 "$pipeline_py" && echo -n "Output: " && ls "$pipeline_py.yaml"; \
    done

FROM registry.access.redhat.com/ubi9/ubi-minimal:9.5

WORKDIR /bin

COPY --from=builder /opt/app-root/src/backend/src/apiserver/config/ /config
COPY --from=builder /bin/apiserver /bin/apiserver

COPY --from=compiler /samples/ /samples/
RUN chmod +x /bin/apiserver

USER root

# Adding CA certificate so API server can download pipeline through URL and wget is used for liveness/readiness probe command
RUN microdnf install -y ca-certificates wget

USER 1001

# Expose apiserver port
EXPOSE 8888

# Start the apiserver
CMD ["/bin/sh", "-c", "/bin/apiserver --config=/config --sampleconfig=/config/sample_config.json -logtostderr=true --logLevel=${LOG_LEVEL}"]
