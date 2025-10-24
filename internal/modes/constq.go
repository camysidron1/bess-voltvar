package modes

import "github.com/example/bess-voltvar/internal/config"

func QStarConstQ(cfg *config.Config) float64 {
	return cfg.QSet
}
