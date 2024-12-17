package cli

import (
	"flag"
	"fmt"
	"os"
)

type Config struct {
	Domain   string
	TaskList string
	Debug    bool
}

func ParseFlags() (*Config, error) {
    cfg := &Config{}

    flag.StringVar(&cfg.Domain, "domain", os.Getenv("SWF_DOMAIN"), "SWF domain")
    flag.StringVar(&cfg.TaskList, "task-list", os.Getenv("SWF_TASK_LIST"), "SWF task list")
    flag.BoolVar(&cfg.Debug, "debug", false, "Enable debug logging")

    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage of swfbq:\n")
        fmt.Fprintf(os.Stderr, "  swfbq [options]\n\n")
        fmt.Fprintf(os.Stderr, "Options:\n")
        flag.PrintDefaults()
    }

    flag.Parse()

    if cfg.Domain == "" {
        return nil, fmt.Errorf("SWF domain is required (use -domain or set SWF_DOMAIN)")
    }
    if cfg.TaskList == "" {
        return nil, fmt.Errorf("SWF task list is required (use -task-list or set SWF_TASK_LIST)")
    }

    return cfg, nil
}
