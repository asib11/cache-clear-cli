package cmd

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"medusa-cache-cli/internal/cache"
	"medusa-cache-cli/internal/config"
	"medusa-cache-cli/internal/ui"
)

const version = "1.0.0"

// Execute is the main entry point — parses top-level subcommand.
func Execute() error {
	ui.Banner()

	if len(os.Args) < 2 {
		printHelp()
		return nil
	}

	subCmd := os.Args[1]
	args := os.Args[2:]

	// Extract --env from args before passing to subcommands
	envFile := ".env"
	filtered := args[:0]
	for i := 0; i < len(args); i++ {
		if args[i] == "--env" && i+1 < len(args) {
			envFile = args[i+1]
			i++
		} else if strings.HasPrefix(args[i], "--env=") {
			envFile = strings.TrimPrefix(args[i], "--env=")
		} else {
			filtered = append(filtered, args[i])
		}
	}

	cfg := config.Load(envFile)

	switch subCmd {
	case "status":
		return runStatus(cfg, filtered)
	case "clear":
		return runClear(cfg, filtered)
	case "version":
		fmt.Printf("medusa-cache-cli v%s\n", version)
		return nil
	case "help", "--help", "-h":
		printHelp()
		return nil
	default:
		ui.Error(fmt.Sprintf("Unknown command: %s", subCmd))
		printHelp()
		return fmt.Errorf("unknown command: %s", subCmd)
	}
}

// ── status ──────────────────────────────────────────────────────────────────

func runStatus(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("status", flag.ExitOnError)
	target := fs.String("target", "all", "What to check: all | redis | nextjs | node | cdn")
	fs.Parse(args) //nolint

	fmt.Println()

	switch *target {
	case "all":
		cache.RedisStatus(cfg.RedisAddr, cfg.RedisPassword)
		fmt.Println()
		cache.NextjsStatus(cfg.NextjsDir)
		fmt.Println()
		cache.NodeModulesStatus(cfg.NextjsDir, cfg.MedusaDir)
		fmt.Println()
		cache.CDNCacheInfo()
	case "redis":
		cache.RedisStatus(cfg.RedisAddr, cfg.RedisPassword)
	case "nextjs":
		cache.NextjsStatus(cfg.NextjsDir)
	case "node":
		cache.NodeModulesStatus(cfg.NextjsDir, cfg.MedusaDir)
	case "cdn":
		cache.CDNCacheInfo()
	default:
		ui.Error("Unknown target: " + *target)
		fmt.Println("  Valid targets: all | redis | nextjs | node | cdn")
		return fmt.Errorf("unknown target: %s", *target)
	}

	fmt.Println()
	return nil
}

// ── clear ───────────────────────────────────────────────────────────────────

func runClear(cfg *config.Config, args []string) error {
	fs := flag.NewFlagSet("clear", flag.ExitOnError)
	target := fs.String("target", "all", "What to clear: all | redis | nextjs | node")
	yes := fs.Bool("yes", false, "Skip confirmation prompt")
	fs.Parse(args) //nolint

	fmt.Println()

	targets := resolveClearTargets(*target)
	if len(targets) == 0 {
		ui.Error("Unknown target: " + *target)
		fmt.Println("  Valid targets: all | redis | nextjs | node")
		return fmt.Errorf("unknown target: %s", *target)
	}

	// Confirmation
	if !*yes {
		fmt.Printf("\n  This will clear: %s\n", strings.Join(targets, ", "))
		ui.Prompt("Are you sure? [y/N]")

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println()
			ui.Warn("Aborted — no caches were cleared.")
			fmt.Println()
			return nil
		}
	}

	totalSteps := len(targets)
	step := 0

	for _, t := range targets {
		step++
		ui.Step(step, totalSteps, "Processing: "+t)
		switch t {
		case "redis":
			cache.ClearRedis(cfg.RedisAddr, cfg.RedisPassword)
		case "nextjs":
			cache.ClearNextjs(cfg.NextjsDir)
		case "node":
			cache.ClearNodeModules(cfg.NextjsDir, cfg.MedusaDir)
		}
	}

	fmt.Println()
	ui.Success("All done! Cache clear complete.")
	fmt.Println()
	return nil
}

func resolveClearTargets(target string) []string {
	switch target {
	case "all":
		return []string{"redis", "nextjs", "node"}
	case "redis", "nextjs", "node":
		return []string{target}
	default:
		return nil
	}
}

// ── help ─────────────────────────────────────────────────────────────────────

func printHelp() {
	fmt.Print(`
  Usage:
    medusa-cache-cli <command> [flags]

  Commands:
    status    Show cache status and stats
    clear     Clear caches
    version   Print version
    help      Show this help

  Flags (all commands):
    --env <path>      Path to .env file (default: .env)

  Status flags:
    --target <name>   all | redis | nextjs | node | cdn   (default: all)

  Clear flags:
    --target <name>   all | redis | nextjs | node         (default: all)
    --yes             Skip confirmation prompt

  Examples:
    medusa-cache-cli status
    medusa-cache-cli status --target redis
    medusa-cache-cli clear
    medusa-cache-cli clear --target nextjs --yes
    medusa-cache-cli clear --env /path/to/.env --yes

  Config (.env or env vars):
    REDIS_URL         Redis address (default: localhost:6379)
    REDIS_PASSWORD    Redis password (optional)
    NEXTJS_DIR        Path to Next.js project (default: ./storefront)
    MEDUSA_DIR        Path to Medusa project  (default: ./backend)

`)
}
