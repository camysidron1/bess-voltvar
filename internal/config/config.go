package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version  int    `yaml:"version"`
	Mode     string `yaml:"mode"` // VOLT_VAR | CONST_PF | CONST_Q | REMOTE
	PF       float64 `yaml:"pf_setpoint"`
	QSet     float64 `yaml:"q_setpoint_mvar"`

	VoltVar struct {
		VDeadbandPU [2]float64   `yaml:"v_deadband_pu"`
		VQPointsPU  [][2]float64 `yaml:"v_q_points_pu"` // [[v_pu,q_pu],...]
	} `yaml:"volt_var"`

	Limits struct {
		SRatingMVA       float64 `yaml:"s_rating_mva"`
		QMaxAbsMVAr      float64 `yaml:"q_max_abs_mvar"`
		QRampMVArPerS    float64 `yaml:"q_ramp_mvar_per_s"`
		VLpfTauS         float64 `yaml:"v_lpf_tau_s"`
	} `yaml:"limits"`

	Fallbacks struct {
		RemoteTimeoutS int    `yaml:"remote_timeout_s"`
		DefaultMode    string `yaml:"default_mode"`
	} `yaml:"fallbacks"`
}

func LoadFromFile(path string) (*Config, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) Validate() error {
	if c.Mode == "" {
		return fmt.Errorf("mode is required")
	}
	if c.Limits.SRatingMVA <= 0 || c.Limits.QMaxAbsMVAr <= 0 {
		return fmt.Errorf("invalid S rating or Qmax")
	}
	if len(c.VoltVar.VQPointsPU) < 2 {
		return fmt.Errorf("volt_var.v_q_points_pu requires at least 2 points")
	}
	return nil
}
