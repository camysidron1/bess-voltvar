package modes

import "github.com/example/bess-voltvar/internal/config"

// QStarVoltVar computes the raw Q* per piecewise-linear Q(V) curve in pu scaled by Qmax.
func QStarVoltVar(cfg *config.Config, v_pu float64) float64 {
	pts := cfg.VoltVar.VQPointsPU
	if v_pu <= pts[0][0] {
		return pts[0][1] * cfg.Limits.QMaxAbsMVAr
	}
	if v_pu >= pts[len(pts)-1][0] {
		return pts[len(pts)-1][1] * cfg.Limits.QMaxAbsMVAr
	}
	for i := 0; i < len(pts)-1; i++ {
		v1, q1 := pts[i][0], pts[i][1]
		v2, q2 := pts[i+1][0], pts[i+1][1]
		if v_pu >= v1 && v_pu <= v2 {
			t := (v_pu - v1) / (v2 - v1)
			qpu := q1 + t*(q2-q1)
			return qpu * cfg.Limits.QMaxAbsMVAr
		}
	}
	return 0
}
