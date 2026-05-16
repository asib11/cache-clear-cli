# medusa-cache-cli

A zero-dependency Golang CLI to manage all caches in your **Medusa v2 + Next.js** project from one place.

## Cache Coverage

| Cache | Location |
|-------|----------|
| **Redis** | Medusa v2 Docker Redis (TCP) |
| **Next.js** | `.next/cache/` directory |
| **Node modules** | `node_modules/.cache/`, `.turbo/` |
| **CDN/Browser** | Info only (clear via your CDN provider) |

## Build & Install

```bash
cd medusa-cache-cli

# Build binary
go build -o medusa-cache-cli .

# Install system-wide (Linux/Mac)
sudo mv medusa-cache-cli /usr/local/bin/

# Or use Makefile
make build
make install
```

## Config

Copy the example env file and adjust to your setup:

```bash
cp .env.example .env
```

```env
# Redis — your docker-compose maps 6380:6379, so from host use localhost:6380
REDIS_URL=localhost:6380
REDIS_PASSWORD=

# Paths to your projects (relative or absolute)
NEXTJS_DIR=../storefront
MEDUSA_DIR=../backend
```

## Usage

### Status

```bash
# All caches
medusa-cache-cli status

# Specific target
medusa-cache-cli status --target redis
medusa-cache-cli status --target nextjs
medusa-cache-cli status --target node
medusa-cache-cli status --target cdn
```

### Clear

```bash
# Clear all (asks for confirmation)
medusa-cache-cli clear

# Clear all without confirmation
medusa-cache-cli clear --yes

# Clear specific cache
medusa-cache-cli clear --target redis --yes
medusa-cache-cli clear --target nextjs --yes
medusa-cache-cli clear --target node --yes
```

### Custom .env path

```bash
medusa-cache-cli status --env /path/to/custom.env
medusa-cache-cli clear --env /path/to/custom.env --yes
```

## Shell Alias (Recommended)

Instead of typing the full command with `--env` every time, add an alias or function to your shell.

### Option 1 — Simple alias

Add to `~/.bashrc` or `~/.zshrc`:

```bash
alias mcc="medusa-cache-cli --env /root/node-sunnah/sfu-backend/medusa-cache-cli/.env"
```

Reload:
```bash
source ~/.bashrc
```

Usage:
```bash
mcc status
mcc clear --yes
```

### Option 2 — Function (supports extra flags)

Add to `~/.bashrc` or `~/.zshrc`:

```bash
mcc() {
  cmd="$1"
  if [ -z "$cmd" ]; then
    echo "Usage: mcc <command> [flags]" >&2
    return 1
  fi
  shift
  medusa-cache-cli "$cmd" --env /root/node-sunnah/sfu-backend/medusa-cache-cli/.env "$@"
}
```

Reload:
```bash
source ~/.bashrc
```

Usage:
```bash
mcc status
mcc status --target redis
mcc clear --yes
mcc clear --target nextjs --yes
```

> The function version is recommended — it correctly handles extra flags like `--target` and `--yes`.

## Docker Compose

Your Redis setup:

```yaml
redis:
  image: redis:7-alpine
  ports:
    - "6380:6379"   # host port 6380 → container port 6379
```

So in `.env` use:
```env
REDIS_URL=localhost:6380
```

The CLI also understands the `redis://` URL format used inside Docker:
```env
REDIS_URL=redis://redis:6379   # works too
```

## Project Structure

```
medusa-cache-cli/
├── main.go
├── cmd/
│   └── root.go           # CLI entry — status, clear commands
├── internal/
│   ├── cache/
│   │   └── cache.go      # Redis flush, Next.js & node cache removal
│   ├── config/
│   │   └── config.go     # .env loader + Redis URL parser
│   ├── redis/
│   │   └── client.go     # Raw TCP Redis client (RESP protocol, no deps)
│   └── ui/
│       └── ui.go         # ANSI color terminal output
├── .env.example
├── Makefile
└── go.mod
```

> **Zero external dependencies** — only Go standard library is used.
