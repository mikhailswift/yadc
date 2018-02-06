package cache

import (
	"time"
)

type action int

const (
	//Failed indicates the attempted action failed to complete successfully.  This should be accompanied by an appropriate error
	Failed action = iota
	//Created indicates the attempted action resulted in a new key being set
	Created action = iota
	//Updated indicates the attempted action resulted in a key being updated
	Updated action = iota
	//Deleted indicates the attempted action resulted in a key being unset
	Deleted action = iota
	//Retrieved indicates the attempted action returning a value
	Retrieved action = iota
)

type Result struct {
	Action action
	n      node
	Err    error
}

func (r Result) GetValue() string {
	return r.n.value
}

func (r Result) GetKey() string {
	return r.n.key
}

func (r Result) GetCreatedTime() time.Time {
	return r.n.created
}
