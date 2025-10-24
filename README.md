# bess-voltvar
A production-style Volt/VAR (Q–V) controller microservice for Battery Energy Storage Systems (BESS) MIC layer.

Implements a deterministic control loop with modes: `VOLT_VAR`, `CONST_PF`, `CONST_Q`, `REMOTE`. 
Enforces inverter capability limits, ramp/slew, interlocks, and exposes a small REST API.

## Quick start
```bash
go mod tidy
go build ./cmd/voltvard
./voltvard -config ./configs/site.example.yaml
```

## REST
- `GET  /healthz`
- `GET  /v1/status`
- `POST /v1/mode`     → `{ "mode": "VOLT_VAR|CONST_PF|CONST_Q|REMOTE" }`
- `POST /v1/remote`   → `{ "q_set_mvar": 0.3, "ttl_s": 5 }`
- `POST /v1/config`   → body: full YAML; atomically swaps after schema validation

## Layout
```
cmd/voltvard/        # main entrypoint
internal/controller/ # control loop & state machine
internal/config/     # config schema + parsing + reload
internal/io/         # measurements in, PCS sink (interfaces + local stubs)
internal/modes/      # strategies for each operating mode
internal/safety/     # interlocks, guards
internal/telem/      # metrics & logging
pkg/api/             # REST API
deploy/              # docker & k8s
configs/             # example config(s)
tests/               # unit & scenario scaffolding
```
