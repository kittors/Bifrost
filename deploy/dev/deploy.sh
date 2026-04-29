#!/usr/bin/env sh
set -eu

ROOT_DIR="$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)"
STATE_DIR="${BIFROST_DEPLOY_STATE_DIR:-/opt/bifrost-dev/shared}"
ENV_FILE="${STATE_DIR}/dev.env"
COMPOSE_FILE="${ROOT_DIR}/deploy/dev/docker-compose.yml"
PROJECT_NAME="bifrost-remote-dev"

generate_secret() {
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex 32
    return
  fi

  dd if=/dev/urandom bs=32 count=1 2>/dev/null | od -An -tx1 | tr -d ' \n'
}

ensure_env_key() {
  key="$1"
  value="$2"

  if ! grep -q "^${key}=" "${ENV_FILE}" 2>/dev/null; then
    printf '%s=%s\n' "${key}" "${value}" >>"${ENV_FILE}"
  fi
}

upsert_env_key() {
  key="$1"
  value="$2"

  if grep -q "^${key}=" "${ENV_FILE}" 2>/dev/null; then
    temp_file="$(mktemp)"
    awk -v key="${key}" -v value="${value}" '
      BEGIN { replaced = 0 }
      $0 ~ "^" key "=" {
        print key "=" value
        replaced = 1
        next
      }
      { print }
      END {
        if (replaced == 0) {
          print key "=" value
        }
      }
    ' "${ENV_FILE}" >"${temp_file}"
    cat "${temp_file}" >"${ENV_FILE}"
    rm -f "${temp_file}"
    return
  fi

  printf '%s=%s\n' "${key}" "${value}" >>"${ENV_FILE}"
}

ensure_env_file() {
  mkdir -p "${STATE_DIR}"

  umask 077
  touch "${ENV_FILE}"
  ensure_env_key "BIFROST_DEV_POSTGRES_PASSWORD" "$(generate_secret)"
  ensure_env_key "BIFROST_DEV_TOKEN_SECRET" "$(generate_secret)"
  upsert_env_key "BIFROST_DEV_GATEWAY_BIND" "${BIFROST_DEV_GATEWAY_BIND:-0.0.0.0}"
  upsert_env_key "BIFROST_DEV_GATEWAY_PORT" "${BIFROST_DEV_GATEWAY_PORT:-18080}"
}

compose() {
  docker compose \
    --project-name "${PROJECT_NAME}" \
    --env-file "${ENV_FILE}" \
    -f "${COMPOSE_FILE}" \
    "$@"
}

gateway_url() {
  awk -F= '/^BIFROST_DEV_GATEWAY_PORT=/ {print "http://127.0.0.1:" $2 "/healthz"}' "${ENV_FILE}"
}

wait_for_gateway() {
  url="$(gateway_url)"
  attempt=1
  while [ "${attempt}" -le 30 ]; do
    if wget -qO- "${url}" >/dev/null 2>&1; then
      return
    fi
    attempt=$((attempt + 1))
    sleep 2
  done

  echo "gateway health check failed at ${url}" >&2
  compose logs gateway >&2
  exit 1
}

assert_private_service() {
  service_name="$1"
  container_id="$(compose ps -q "${service_name}")"
  if [ -z "${container_id}" ]; then
    echo "${service_name} container is not running" >&2
    exit 1
  fi

  if docker inspect --format '{{range $port, $bindings := .NetworkSettings.Ports}}{{range $bindings}}{{println .HostPort}}{{end}}{{end}}' "${container_id}" | grep -q .; then
    echo "${service_name} unexpectedly exposes a public port" >&2
    exit 1
  fi
}

ensure_env_file
compose build gateway mock-gitlab mock-jenkins mock-docs mock-internal-admin
compose up -d postgres mock-gitlab mock-jenkins mock-docs mock-internal-admin
compose run --rm migrate up
compose run --rm migrate seed
compose up -d gateway
wait_for_gateway

assert_private_service postgres
assert_private_service mock-gitlab
assert_private_service mock-jenkins
assert_private_service mock-docs
assert_private_service mock-internal-admin

compose ps
