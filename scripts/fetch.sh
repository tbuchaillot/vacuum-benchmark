#!/usr/bin/env bash
set -euo pipefail

# SHA-pinned OpenAPI spec fetch. Idempotent.
# Edit the SHAs below to update; re-run `make fetch`.

PETSTORE_SHA="6e0ed4df20408bd0d5ba15fa5ac53964d9a205aa"
DO_SHA="38172fa5fe619624c89e5b12a0bdfef90e766277"
GITHUB_SHA="ed0da17e9a4a1937de02bc311a9286009ccb7e2b"

PETSTORE_URL="https://raw.githubusercontent.com/OAI/OpenAPI-Specification/${PETSTORE_SHA}/_archive_/schemas/v3.0/pass/petstore.yaml"
DO_URL="https://raw.githubusercontent.com/digitalocean/openapi/${DO_SHA}/specification/DigitalOcean-public.v2.yaml"
GITHUB_URL="https://raw.githubusercontent.com/github/rest-api-description/${GITHUB_SHA}/descriptions/api.github.com/api.github.com.yaml"

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
OUT="${REPO_ROOT}/testdata/openapis"
mkdir -p "${OUT}"

fetch() {
    local url="$1"
    local dest="$2"
    echo "fetching ${url} -> ${dest}"
    curl --fail --silent --show-error --location -o "${dest}" "${url}"
    echo "  size: $(wc -c < "${dest}") bytes"
}

fetch "${PETSTORE_URL}" "${OUT}/petstore.yaml"
fetch "${DO_URL}"       "${OUT}/digitalocean.yaml"
fetch "${GITHUB_URL}"   "${OUT}/github.yaml"

echo "done. fetched specs:"
ls -la "${OUT}"
