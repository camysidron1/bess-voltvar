package safety

type State string

const (
	StateDisabled State = "DISABLED"
	StateArmed    State = "ARMED"
	StateActive   State = "ACTIVE"
	StateFaulted  State = "FAULTED"
)

type StateMachine struct {
	state State
}

func NewStateMachine() StateMachine {
	return StateMachine{state: StateDisabled}
}

func (s *StateMachine) Arm() { s.state = StateArmed }
func (s *StateMachine) Activate() { s.state = StateActive }
func (s *StateMachine) Fault() { s.state = StateFaulted }
func (s *StateMachine) Disable() { s.state = StateDisabled }
func (s *StateMachine) State() State { return s.state }
