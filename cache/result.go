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
	//RetrievedTTL indicated the attempted action is returning a TTL
	RetrievedTTL action = iota
	//Cleared indicates the attempted action cleared the cache
	Cleared action = iota
)

//Result represents a result from the cache table.  Err will be nil when the action was successful and an action of Failed will always have a non-nill Err
type Result struct {
	Action action
	ttl    time.Duration
	n      node
	Err    error
}

//GetValue gets the value of the Node from the cache
func (r Result) GetValue() string {
	return r.n.value
}

//GetKey gets the key of the Node from the cache
func (r Result) GetKey() string {
	return r.n.key
}

//GetCreatedTime gets the Created Time of the Node from the cache
func (r Result) GetCreatedTime() time.Time {
	return r.n.created
}

//GetTTL gets the current TTL.  Only valid for GetTTL calls
func (r Result) GetTTL() time.Duration {
	return r.ttl
}
