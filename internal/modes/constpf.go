package modes

import (
	"math"

	"github.com/example/bess-voltvar/internal/config"
)

func QStarConstPF(cfg *config.Config, p_mw float64) float64 {
	pf := cfg.PF
	if pf < 0.1 {
		pf = 0.1
	}
	if pf > 0.9999 {
		pf = 0.9999
	}
	return p_mw * math.Tan(math.Acos(pf))
}
