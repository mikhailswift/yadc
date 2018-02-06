package cache

import (
	"time"
)

type node struct {
	key     string
	value   string
	created time.Time
}

func newRecord(k, v string) *node {
	return &node{
		key:     k,
		value:   v,
		created: time.Now().UTC(),
	}
}
