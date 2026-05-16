package cache

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"medusa-cache-cli/internal/redis"
	"medusa-cache-cli/internal/ui"
)

// RedisStatus shows Redis cache info.
func RedisStatus(addr, password string) error {
	ui.Section("Redis Cache  (Medusa v2 / Docker)")

	client, err := redis.New(addr, password)
	if err != nil {
		ui.Error(fmt.Sprintf("Cannot connect: %v", err))
		return err
	}
	defer client.Close()
	ui.Success("Connected to " + addr)

	// DB size
	size, err := client.DBSize()
	if err != nil {
		ui.Warn("Could not get DB size: " + err.Error())
	} else {
		ui.Stat("Total keys", strconv.Itoa(size), "keys")
	}

	// INFO server
	info, err := client.Info("server")
	if err == nil {
		for _, line := range strings.Split(info, "\n") {
			line = strings.TrimSpace(line)
			if kv := strings.SplitN(line, ":", 2); len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				val := strings.TrimSpace(kv[1])
				switch key {
				case "redis_version":
					ui.Info("Redis version", val)
				case "uptime_in_seconds":
					sec, _ := strconv.Atoi(val)
					ui.Info("Uptime", formatDuration(sec))
				case "used_memory_human":
					ui.Stat("Memory used", val, "")
				case "maxmemory_human":
					if val != "0B" {
						ui.Stat("Max memory", val, "")
					}
				}
			}
		}
	}

	// INFO stats — hits/misses
	stats, err := client.Info("stats")
	if err == nil {
		hits, misses := "?", "?"
		for _, line := range strings.Split(stats, "\n") {
			line = strings.TrimSpace(line)
			if kv := strings.SplitN(line, ":", 2); len(kv) == 2 {
				switch strings.TrimSpace(kv[0]) {
				case "keyspace_hits":
					hits = strings.TrimSpace(kv[1])
				case "keyspace_misses":
					misses = strings.TrimSpace(kv[1])
				}
			}
		}
		ui.Stat("Cache hits", hits, "")
		ui.Stat("Cache misses", misses, "")
	}

	// Medusa-specific key prefixes
	ui.Divider()
	prefixes := []string{"medusa:*", "product:*", "cart:*", "session:*", "price:*"}
	for _, pat := range prefixes {
		keys, err := client.Keys(pat)
		if err != nil || len(keys) == 0 {
			continue
		}
		ui.Stat("  "+pat, strconv.Itoa(len(keys)), "keys")
	}

	return nil
}

// NextjsStatus shows Next.js .next/cache info.
func NextjsStatus(nextjsDir string) error {
	ui.Section("Next.js Cache  (.next/cache)")

	cacheDir := filepath.Join(nextjsDir, ".next", "cache")
	info, err := os.Stat(cacheDir)
	if err != nil {
		ui.Warn(fmt.Sprintf("Cache dir not found: %s", cacheDir))
		ui.Info("Expected at", cacheDir)
		return nil
	}

	ui.Success("Cache directory found")
	ui.Info("Path", cacheDir)
	ui.Info("Last modified", info.ModTime().Format("2006-01-02 15:04:05"))

	// du -sh for total size
	out, err := exec.Command("du", "-sh", cacheDir).Output()
	if err == nil {
		parts := strings.Fields(string(out))
		if len(parts) > 0 {
			ui.Stat("Total size", parts[0], "")
		}
	}

	// Count subdirs
	entries, err := os.ReadDir(cacheDir)
	if err == nil {
		ui.Stat("Cache entries", strconv.Itoa(len(entries)), "items")
		for _, e := range entries {
			if e.IsDir() {
				sub, _ := os.ReadDir(filepath.Join(cacheDir, e.Name()))
				ui.Info("  └─ "+e.Name()+"/", strconv.Itoa(len(sub))+" files")
			}
		}
	}

	return nil
}

// NodeModulesStatus checks common node_modules caches.
func NodeModulesStatus(nextjsDir, medusaDir string) error {
	ui.Section("Node Modules Cache")

	dirs := []struct{ label, path string }{
		{"Next.js node_modules/.cache", filepath.Join(nextjsDir, "node_modules", ".cache")},
		{"Medusa node_modules/.cache", filepath.Join(medusaDir, "node_modules", ".cache")},
		{"Next.js .turbo", filepath.Join(nextjsDir, ".turbo")},
		{"Medusa .turbo", filepath.Join(medusaDir, ".turbo")},
	}

	found := false
	for _, d := range dirs {
		if _, err := os.Stat(d.path); err == nil {
			found = true
			out, _ := exec.Command("du", "-sh", d.path).Output()
			size := "?"
			if parts := strings.Fields(string(out)); len(parts) > 0 {
				size = parts[0]
			}
			ui.Stat(d.label, size, "")
		}
	}
	if !found {
		ui.Info("Status", "No node_modules caches found")
	}
	return nil
}

// ClearRedis flushes all Redis keys (FLUSHDB).
func ClearRedis(addr, password string) error {
	ui.Section("Clearing Redis Cache")

	client, err := redis.New(addr, password)
	if err != nil {
		ui.Error("Cannot connect: " + err.Error())
		return err
	}
	defer client.Close()

	// Count before
	size, _ := client.DBSize()
	ui.Info("Keys before flush", strconv.Itoa(size))

	start := time.Now()
	if err := client.FlushDB(); err != nil {
		ui.Error("FlushDB failed: " + err.Error())
		return err
	}

	after, _ := client.DBSize()
	elapsed := time.Since(start)
	ui.Success(fmt.Sprintf("Redis flushed in %dms — %d keys removed", elapsed.Milliseconds(), size-after))
	return nil
}

// ClearNextjs removes the .next/cache directory.
func ClearNextjs(nextjsDir string) error {
	ui.Section("Clearing Next.js Cache")

	cacheDir := filepath.Join(nextjsDir, ".next", "cache")
	if _, err := os.Stat(cacheDir); err != nil {
		ui.Warn("Cache dir not found — nothing to clear: " + cacheDir)
		return nil
	}

	// Size before
	out, _ := exec.Command("du", "-sh", cacheDir).Output()
	size := "?"
	if parts := strings.Fields(string(out)); len(parts) > 0 {
		size = parts[0]
	}
	ui.Info("Size to remove", size)

	if err := os.RemoveAll(cacheDir); err != nil {
		ui.Error("Failed to remove: " + err.Error())
		return err
	}
	ui.Success("Removed: " + cacheDir)
	return nil
}

// ClearNodeModules removes node_modules/.cache and .turbo dirs.
func ClearNodeModules(nextjsDir, medusaDir string) error {
	ui.Section("Clearing Node Modules Cache")

	dirs := []struct{ label, path string }{
		{"Next.js node_modules/.cache", filepath.Join(nextjsDir, "node_modules", ".cache")},
		{"Medusa node_modules/.cache", filepath.Join(medusaDir, "node_modules", ".cache")},
		{"Next.js .turbo", filepath.Join(nextjsDir, ".turbo")},
		{"Medusa .turbo", filepath.Join(medusaDir, ".turbo")},
	}

	cleared := 0
	for _, d := range dirs {
		if _, err := os.Stat(d.path); err != nil {
			continue
		}
		if err := os.RemoveAll(d.path); err != nil {
			ui.Error(fmt.Sprintf("Failed %s: %v", d.label, err))
		} else {
			ui.Success("Cleared: " + d.label)
			cleared++
		}
	}

	if cleared == 0 {
		ui.Info("Result", "No caches found to clear")
	}
	return nil
}

// SetCDNCacheHeaders prints a reminder since CDN cache is managed externally.
func CDNCacheInfo() {
	ui.Section("Browser / CDN Cache Headers")
	ui.Info("Type", "HTTP cache headers (server-side)")
	ui.Info("Where to configure", "Next.js headers() in next.config.js")
	ui.Info("CDN purge", "Use your CDN provider's API (Cloudflare, Vercel, etc.)")
	ui.Warn("This CLI cannot clear browser or CDN caches directly.")
	ui.Info("Tip", "Set Cache-Control: no-store for dev, use SWR/ISR for prod")
}

func formatDuration(sec int) string {
	d := time.Duration(sec) * time.Second
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	mins := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, mins)
	}
	return fmt.Sprintf("%dm", mins)
}
