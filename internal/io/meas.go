package io

import "sync"

type Measurements struct {
	VPCC_kV      float64
	VNom_kV      float64
	P_MW         float64
	Q_MVAr       float64
	GridTied     bool
	ThermalDerate float64
}

type MeasurementsProvider interface {
	Get() Measurements
	Set(m Measurements)
}

type LocalMeasurements struct {
	mu  sync.RWMutex
	val Measurements
}

func NewLocalMeasurements() *LocalMeasurements {
	// Some sane defaults
	return &LocalMeasurements{
		val: Measurements{
			VPCC_kV: 13.8, VNom_kV: 13.8,
			P_MW: 0.0, Q_MVAr: 0.0,
			GridTied: true, ThermalDerate: 1.0,
		},
	}
}

func (l *LocalMeasurements) Get() Measurements {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.val
}

func (l *LocalMeasurements) Set(m Measurements) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.val = m
}
