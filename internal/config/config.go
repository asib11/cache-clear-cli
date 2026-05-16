package config

import (
	"bufio"
	"os"
	"strings"
)

// Config holds all settings for the CLI.
type Config struct {
	RedisAddr     string
	RedisPassword string
	RedisDB       string
	NextjsDir     string
	MedusaDir     string
}

// Load tries to read config from .env file, then falls back to env vars.
func Load(envFile string) *Config {
	vars := map[string]string{}

	if f, err := os.Open(envFile); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
				vars[key] = val
			}
		}
	}

	get := func(key, fallback string) string {
		if v, ok := vars[key]; ok && v != "" {
			return v
		}
		if v := os.Getenv(key); v != "" {
			return v
		}
		return fallback
	}

	// REDIS_URL supports:
	//   redis://redis:6379         (Docker Compose internal)
	//   redis://:pass@host:6379    (with password)
	//   localhost:6380             (host-mapped port, plain host:port)
	// Your docker-compose: ports 6380:6379 → from host use localhost:6380
	redisURL := get("REDIS_URL", "localhost:6380")

	return &Config{
		RedisAddr:     parseRedisAddr(redisURL),
		RedisPassword: parseRedisPassword(redisURL, get("REDIS_PASSWORD", "")),
		RedisDB:       get("REDIS_DB", "0"),
		NextjsDir:     get("NEXTJS_DIR", "./storefront"),
		MedusaDir:     get("MEDUSA_DIR", "./backend"),
	}
}

// parseRedisAddr extracts host:port from a Redis URL or plain host:port.
//   "redis://redis:6379"       → "redis:6379"
//   "redis://localhost:6380"   → "localhost:6380"
//   "redis://:pass@host:6379"  → "host:6379"
//   "localhost:6380"           → "localhost:6380"
func parseRedisAddr(rawURL string) string {
	if !strings.HasPrefix(rawURL, "redis://") {
		return rawURL
	}
	rest := strings.TrimPrefix(rawURL, "redis://")
	if at := strings.LastIndex(rest, "@"); at != -1 {
		rest = rest[at+1:]
	}
	if slash := strings.Index(rest, "/"); slash != -1 {
		rest = rest[:slash]
	}
	if q := strings.Index(rest, "?"); q != -1 {
		rest = rest[:q]
	}
	if rest == "" {
		return "localhost:6380"
	}
	return rest
}

// parseRedisPassword extracts password from redis://:password@host URL.
// Explicit REDIS_PASSWORD env var takes priority.
func parseRedisPassword(rawURL, explicit string) string {
	if explicit != "" {
		return explicit
	}
	if !strings.HasPrefix(rawURL, "redis://") {
		return ""
	}
	rest := strings.TrimPrefix(rawURL, "redis://")
	at := strings.LastIndex(rest, "@")
	if at == -1 {
		return ""
	}
	userinfo := rest[:at]
	if colon := strings.Index(userinfo, ":"); colon != -1 {
		return userinfo[colon+1:]
	}
	return userinfo
}
