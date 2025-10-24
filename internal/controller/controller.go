package controller

import (
	"context"
	"math"
	"sync"
	"time"

	"github.com/example/bess-voltvar/internal/config"
	"github.com/example/bess-voltvar/internal/io"
	"github.com/example/bess-voltvar/internal/modes"
	"github.com/example/bess-voltvar/internal/safety"
	"github.com/example/bess-voltvar/internal/telem"
)

type Status struct {
	Mode         string  `json:"mode"`
	VPUPresent   float64 `json:"v_pu"`
	QSetMVAr     float64 `json:"q_set_mvar"`
	QCapMVAr     float64 `json:"q_cap_mvar"`
	Clamped      bool    `json:"clamped"`
	Derate       float64 `json:"derate"`
	GridTied     bool    `json:"grid_tied"`
	LastTickMs   int64   `json:"last_tick_ms"`

	// §5.4.1 Monitoring rules — expose measured values at interconnection point
	Voltage_kV   float64 `json:"voltage_kv"`
	Frequency_Hz float64 `json:"frequency_hz"`
	Active_P_MW  float64 `json:"active_power_mw"`
}

type Controller struct {
	mu     sync.RWMutex
	cfg    *config.Config
	meas   io.MeasurementsProvider
	pcs    io.PCSSink
	tick   time.Duration

	state  safety.StateMachine

	// fast path vars
	vFilt     float64
	qCmd      float64
	lastTick  time.Time
	remoteTS  time.Time
	remoteQ   float64
}

func NewController(cfg *config.Config, m io.MeasurementsProvider, p io.PCSSink, tick time.Duration) *Controller {
	return &Controller{
		cfg:   cfg,
		meas:  m,
		pcs:   p,
		tick:  tick,
		state: safety.NewStateMachine(),
	}
}

func (c *Controller) Run(ctx context.Context) {
	t := time.NewTicker(c.tick)
	defer t.Stop()

	c.state.Arm()

	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			c.step()
		}
	}
}

func (c *Controller) step() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	dt := c.tick.Seconds()
	if c.lastTick.IsZero() {
		c.lastTick = now.Add(-c.tick)
	}

	me := c.meas.Get()
	// LPF voltage per config
	a := dt / (c.cfg.Limits.VLpfTauS + dt)
	if c.vFilt == 0 {
		c.vFilt = me.VPCC_kV / max(me.VNom_kV, 1e-6)
	} else {
		target := me.VPCC_kV / max(me.VNom_kV, 1e-6)
		c.vFilt += a * (target - c.vFilt)
	}

	// determine mode
	mode := c.cfg.Mode
	// handle remote timeout
	if mode == "REMOTE" {
		if time.Since(c.remoteTS) > time.Duration(c.cfg.Fallbacks.RemoteTimeoutS)*time.Second {
			mode = c.cfg.Fallbacks.DefaultMode
			telem.Log().Printf("remote timeout → fallback to %s", mode)
		}
	}

	// raw q*
	var qStar float64
	switch mode {
	case "VOLT_VAR":
		qStar = modes.QStarVoltVar(c.cfg, c.vFilt)
	case "CONST_PF":
		qStar = modes.QStarConstPF(c.cfg, me.P_MW)
	case "CONST_Q":
		qStar = modes.QStarConstQ(c.cfg)
	case "REMOTE":
		qStar = c.remoteQ
	default:
		qStar = 0
	}

	// capability: |Q| <= min(Qmax, sqrt(S^2 - P^2))
	qCap := math.Min(c.cfg.Limits.QMaxAbsMVAr, math.Sqrt(max(c.cfg.Limits.SRatingMVA*c.cfg.Limits.SRatingMVA - me.P_MW*me.P_MW, 0)))
	if qStar > qCap {
		qStar = qCap
	} else if qStar < -qCap {
		qStar = -qCap
	}
	clamped := math.Abs(qStar) >= qCap-1e-6

	// interlocks/derates
	if !me.GridTied {
		qStar = 0
	}
	qStar *= clamp(me.ThermalDerate, 0, 1)

	// ramp/slew
	r := c.cfg.Limits.QRampMVArPerS * dt
	qMin := c.qCmd - r
	qMax := c.qCmd + r
	if qStar > qMax {
		qStar = qMax
	} else if qStar < qMin {
		qStar = qMin
	}
	c.qCmd = qStar

	// write to PCS (non-blocking)
	c.pcs.SetReactivePower(c.qCmd)

	c.lastTick = now
	_ = clamped // for future telemetry
}

func (c *Controller) Status() Status {
	c.mu.RLock()
	defer c.mu.RUnlock()
	st := Status{
		Mode:       c.cfg.Mode,
		VPUPresent: c.vFilt,
		QSetMVAr:   c.qCmd,
		QCapMVAr:   math.Min(c.cfg.Limits.QMaxAbsMVAr, math.Sqrt(max(c.cfg.Limits.SRatingMVA*c.cfg.Limits.SRatingMVA - c.meas.Get().P_MW*c.meas.Get().P_MW, 0))),
		Clamped:    false,
		Derate:     c.meas.Get().ThermalDerate,
		GridTied:   c.meas.Get().GridTied,
		LastTickMs: c.tick.Milliseconds(),
	}
	me := c.meas.get()
	st.Voltage_kV = me.VPCC_kV
	st.Frequency_Hz = me.F_Hz
	st.Active_P_MW = me.P_MW
	return st
}

func (c *Controller) UpdateConfig(newCfg *config.Config) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cfg = newCfg
}

func (c *Controller) SetMode(m string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cfg.Mode = m
}

func (c *Controller) SetRemote(q float64, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.remoteQ = q
	c.remoteTS = time.Now()
	_ = ttl // ttl handled via timestamp
}

func max(a, b float64) float64 { if a > b { return a }; return b }
func clamp(x, lo, hi float64) float64 {
	if x < lo { return lo }
	if x > hi { return hi }
	return x
}
