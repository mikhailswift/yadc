package cache

import (
	"fmt"
)

type ErrKeyNotFound string

func (e ErrKeyNotFound) Error() string {
	return fmt.Sprintf("Could not find key: %+v", e)
}
