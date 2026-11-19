#!/bin/bash
#
# Copyright 2023 kubeflow.org
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# source: https://raw.githubusercontent.com/open-toolchain/commons/master/scripts/check_registry.sh

# Remove the x if you need no print out of each command
set -e

REGISTRY="${REGISTRY:-kind-registry:5000}"
echo "REGISTRY=$REGISTRY"
TAG="${TAG:-latest}"
EXIT_CODE=0

docker system prune -a -f

APPS=("apiserver" "persistenceagent" "scheduledworkflow" "driver" "launcher")
DOCKERFILES=("backend/Dockerfile" "backend/Dockerfile.persistenceagent" "backend/Dockerfile.scheduledworkflow" "backend/Dockerfile.driver" "backend/Dockerfile.launcher")

for i in "${!APPS[@]}"; do
  app="${APPS[$i]}"
  dockerfile="${DOCKERFILES[$i]}"
  echo "🔨 Building ${app}..."
  docker build --progress=plain -t "${REGISTRY}/${app}:${TAG}" -f "${dockerfile}" . && docker push "${REGISTRY}/${app}:${TAG}" || EXIT_CODE=$?
  if [[ $EXIT_CODE -ne 0 ]]; then
    echo "Failed to build/push ${app} image."
    exit $EXIT_CODE
  fi
  # Remove local image to free disk space
  docker image rm "${REGISTRY}/${app}:${TAG}" || true
done

# clean up intermittent build caches to free up disk space
docker system prune -a -f
