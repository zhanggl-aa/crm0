# YAML Config Multi-Environment Design

## Goal

Replace the current env-var-only config system with YAML config files, supporting dev/test/prod environments via a `-f` flag, while keeping env var override capability.

## Config File Structure

Three YAML files in `backend/configs/`:

- `config.dev.yaml` — development (localhost, dummy secrets)
- `config.test.yaml` — test environment
- `config.prod.yaml` — production (sensitive values as placeholders, overridden by env vars)

YAML format with nested sections:

```yaml
server:
  port: 3000

database:
  host: localhost
  port: 5432
  user: postgres
  password: "123456"
  name: crm0_db

jwt:
  secret: crm0-secret-key
  expiry_hours: 72

algorithm:
  service_url: http://localhost:8001

redis:
  url: localhost:6379

stripe:
  secret_key: ""
  webhook_secret: ""

frontend:
  url: http://localhost:5173

encryption:
  key: default-encryption-key-change-in-prod
```

## Runtime Flag

```bash
go run cmd/server/main.go -f config.prod.yaml
./server -f config.prod.yaml
```

Default: `configs/config.dev.yaml` when `-f` is not specified.

## Loading Priority

YAML file values loaded first, then environment variables override matching fields.

Env var names remain the same as today: `DB_HOST`, `DB_PORT`, `JWT_SECRET`, etc.

## Implementation Changes

1. Add `gopkg.in/yaml.v3` dependency
2. Remove `github.com/joho/godotenv` dependency
3. Restructure `Config` struct with nested YAML-tagged sub-structs
4. Add `LoadFromFile(path string)` that reads YAML then applies env var overrides
5. Modify `main.go` to parse `-f` flag
6. Keep `DSN()` and `AppAddr()` accessor methods working
7. Create three YAML config files in `backend/configs/`

## Affected Files

- `backend/internal/config/config.go` — major rewrite
- `backend/cmd/server/main.go` — add -f flag parsing
- `backend/go.mod` / `backend/go.sum` — add yaml.v3, remove godotenv
- New: `backend/configs/config.dev.yaml`
- New: `backend/configs/config.test.yaml`
- New: `backend/configs/config.prod.yaml`
