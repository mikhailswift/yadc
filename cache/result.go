package cache

type action int

const (
	Failed    action = iota
	Created   action = iota
	Updated   action = iota
	Deleted   action = iota
	Retrieved action = iota
)

type result struct {
	Action action
	Node   node
	Err    error
}
