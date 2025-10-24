package unit

import (
	"testing"

	"github.com/example/bess-voltvar/internal/config"
	"github.com/example/bess-voltvar/internal/modes"
)

func TestVoltVarInterpolation(t *testing.T) {
	cfg := &config.Config{}
	cfg.Limits.QMaxAbsMVAr = 5
	cfg.VoltVar.VQPointsPU = [][2]float64{
		{0.94, 1.0},
		{0.985, 0.0},
		{1.015, 0.0},
		{1.06, -1.0},
	}
	q := modes.QStarVoltVar(cfg, 0.9625) // halfway 0.94->0.985 => qpu=0.5
	if want, got := 2.5, q; (got - want) > 1e-6 {
		t.Fatalf("want %.3f got %.3f", want, got)
	}
}
