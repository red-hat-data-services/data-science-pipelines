#!/usr/bin/env bash
# Verifies that all Dockerfiles using a golang or go-toolset base image
# specify a Go version consistent with the root go.mod.

set -euo pipefail

REPO_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"

GOMOD_VERSION=$(grep -E '^go [0-9]' "$REPO_ROOT/go.mod" | awk '{print $2}' || true)

if [[ -z "$GOMOD_VERSION" ]]; then
    echo "ERROR: Could not extract Go version from go.mod" >&2
    exit 1
fi

echo "go.mod Go version: $GOMOD_VERSION"

IGNORE_FILE="$REPO_ROOT/.github/scripts/go-version-consistency-ignore"
IGNORED_PATHS=()
if [[ -f "$IGNORE_FILE" ]]; then
    while IFS= read -r line; do
        line="${line%%#*}"
        line="${line#"${line%%[![:space:]]*}"}"
        line="${line%"${line##*[![:space:]]}"}"
        [[ -n "$line" ]] && IGNORED_PATHS+=("$line")
    done < "$IGNORE_FILE"
fi

version_matches() {
    local docker_version="$1" gomod_version="$2"
    local docker_major_minor gomod_major_minor
    docker_major_minor=$(echo "$docker_version" | cut -d. -f1-2)
    gomod_major_minor=$(echo "$gomod_version" | cut -d. -f1-2)
    if [[ "$docker_major_minor" != "$gomod_major_minor" ]]; then
        return 1
    fi
    if [[ "$docker_version" == *.*.* && "$gomod_version" == *.*.* ]]; then
        [[ "$docker_version" != "$gomod_version" ]] && return 1
    fi
    return 0
}

ERRORS=0
CHECKED=0
TOTAL_FOUND=0
FOUND=0

while IFS= read -r dockerfile; do
    relative="${dockerfile#"$REPO_ROOT"/}"
    TOTAL_FOUND=$((TOTAL_FOUND + 1))
    skip=false
    for ignored in "${IGNORED_PATHS[@]}"; do
        if [[ "$relative" == "$ignored" ]]; then
            echo "  SKIP: $relative (ignored)"
            skip=true
            break
        fi
    done
    [[ "$skip" == true ]] && continue
    FOUND=$((FOUND + 1))
    while IFS= read -r line; do
        docker_version=$(echo "$line" | sed -E 's/.*FROM[[:space:]]+(--[^[:space:]]+[[:space:]]+)*(golang|[^[:space:]]*go-toolset):([0-9]+\.[0-9]+(\.[0-9]+)?).*/\3/')

        if [[ ! "$docker_version" =~ ^[0-9]+\.[0-9]+(\.[0-9]+)?$ ]]; then
            echo "ERROR: Could not parse Go version from line in $relative: $line" >&2
            ERRORS=$((ERRORS + 1))
            continue
        fi

        CHECKED=$((CHECKED + 1))

        if ! version_matches "$docker_version" "$GOMOD_VERSION"; then
            echo "MISMATCH: $relative has Go $docker_version, but go.mod requires $GOMOD_VERSION" >&2
            ERRORS=$((ERRORS + 1))
        else
            echo "  OK: $relative (Go $docker_version)"
        fi
    done < <(grep -iE '^FROM[[:space:]]+(--[^[:space:]]+[[:space:]]+)*(golang|[^[:space:]]*go-toolset):' "$dockerfile" || true)
done < <(cd "$REPO_ROOT" && (git ls-files -z '*Dockerfile*' | xargs -0 grep -liE -- 'FROM[[:space:]]+(--[^[:space:]]+[[:space:]]+)*(golang|[^[:space:]]*go-toolset):' | sed "s|^|$REPO_ROOT/|") || true)

echo ""

if [[ $TOTAL_FOUND -eq 0 ]]; then
    echo "ERROR: No Dockerfiles with Go base images found." >&2
    exit 1
fi

if [[ $FOUND -eq 0 ]]; then
    echo "INFO: All $TOTAL_FOUND Dockerfile(s) with Go base images are ignored. Nothing to check."
    exit 0
fi

if [[ $CHECKED -eq 0 ]]; then
    echo "ERROR: Found $FOUND Dockerfile(s) with Go base images, but could not parse any Go version." >&2
    exit 1
fi

if [[ $ERRORS -gt 0 ]]; then
    echo "FAILED: $ERRORS error(s) found when checking Go base image stages against go.mod ($GOMOD_VERSION)." >&2
    exit 1
fi

echo "PASSED: All $CHECKED Go base image stage(s) use Go $GOMOD_VERSION, matching go.mod."
