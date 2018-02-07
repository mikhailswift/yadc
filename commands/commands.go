package commands

import (
	"fmt"
	"time"

	"github.com/mikhailswift/yadc/cache"
)

//Command is the interface all commands that interact with the cache must satisfy
type Command interface {
	Execute(cache.Cacher) cache.Result
	fmt.Stringer
}

//GetCommand is a command that will get a key from the cache.
type GetCommand struct {
	Key string
}

//Execute gets the key from the cache and returns the result
func (cmd GetCommand) Execute(c cache.Cacher) cache.Result {
	return c.Get(cmd.Key)
}

func (cmd GetCommand) String() string {
	return fmt.Sprintf("GET \"%v\"", cmd.Key)
}

//NewGetCommand returns a command that will retrieve a key from the cache.
func NewGetCommand(key string) Command {
	return GetCommand{
		Key: key,
	}
}

//SetCommand is a command that will set a key in the cache to the provided value.  If TTL is > 0 it will also set the TTL for the key.
type SetCommand struct {
	Key   string
	Value string
	TTL   time.Duration
}

//Execute sets the provided key with the provided value.  If TTL is greater than zero it will also set a TTL.
func (cmd SetCommand) Execute(c cache.Cacher) cache.Result {
	return c.Set(cmd.Key, cmd.Value, cmd.TTL)
}

func (cmd SetCommand) String() string {
	if cmd.TTL > 0 {
		return fmt.Sprintf("SET \"%v\" \"%v\" %v", cmd.Key, cmd.Value, int(cmd.TTL.Seconds()))
	}

	return fmt.Sprintf("SET \"%v\" \"%v\"", cmd.Key, cmd.Value)
}

//NewSetCommand returns a new command that will set a key to a value.  If ttl is greater than zero it will also set a TTL.
func NewSetCommand(key, value string, ttl time.Duration) Command {
	return SetCommand{
		Key:   key,
		Value: value,
		TTL:   ttl,
	}
}

//UnsetCommand is a command that will unset a key in the cache.
type UnsetCommand struct {
	Key string
}

//Execute unsets the provided key.
func (cmd UnsetCommand) Execute(c cache.Cacher) cache.Result {
	return c.Unset(cmd.Key)
}

func (cmd UnsetCommand) String() string {
	return fmt.Sprintf("UNSET \"%v\"", cmd.Key)
}

//NewUnsetCommand returns a new command that will unset a key.
func NewUnsetCommand(key string) Command {
	return UnsetCommand{
		Key: key,
	}
}

//GetTTLCommand is a command that will get the TTL for a key in the cache.
type GetTTLCommand struct {
	Key string
}

//Execute gets the TTL for the the provided key.
func (cmd GetTTLCommand) Execute(c cache.Cacher) cache.Result {
	return c.GetTTL(cmd.Key)
}

func (cmd GetTTLCommand) String() string {
	return fmt.Sprintf("GETTTL \"%v\"", cmd.Key)
}

//NewGetTTLCommand returns a new command that will get the TTL for a key.
func NewGetTTLCommand(key string) Command {
	return GetTTLCommand{
		Key: key,
	}
}

//SetTTLCommand is a command that will set the TTL for a key in the cache.
type SetTTLCommand struct {
	Key string
	TTL time.Duration
}

//Execute sets the TTL for the the provided key.
func (cmd SetTTLCommand) Execute(c cache.Cacher) cache.Result {
	return c.SetTTL(cmd.Key, cmd.TTL)
}

func (cmd SetTTLCommand) String() string {
	return fmt.Sprintf("SETTTL \"%v\" %v", cmd.Key, int(cmd.TTL.Seconds()))
}

//NewSetTTLCommand returns a new command that will set the TTL for a key.
func NewSetTTLCommand(key string, ttl time.Duration) Command {
	return SetTTLCommand{
		Key: key,
		TTL: ttl,
	}
}

//UnsetTTLCommand is a command that will unset the TTL for a key in the cache.
type UnsetTTLCommand struct {
	Key string
}

//Execute unsets the TTL for the the provided key.
func (cmd UnsetTTLCommand) Execute(c cache.Cacher) cache.Result {
	return c.SetTTL(cmd.Key, 0*time.Second)
}

func (cmd UnsetTTLCommand) String() string {
	return fmt.Sprintf("UNSETTTL \"%v\"", cmd.Key)
}

//NewUnsetTTLCommand returns a new command that will get the TTL for a key.
func NewUnsetTTLCommand(key string) Command {
	return UnsetTTLCommand{
		Key: key,
	}
}

//UnsetAllCommand is a command that will erase all keys from the cache.
type UnsetAllCommand struct{}

//Execute unsets all the keys in the cache.
func (cmd UnsetAllCommand) Execute(c cache.Cacher) cache.Result {
	return c.UnsetAll()
}

func (cmd UnsetAllCommand) String() string {
	return "UNSETALL"
}

//NewUnsetAllCommand returns a new command that will wipe all the keys in the cache.
func NewUnsetAllCommand() Command {
	return UnsetAllCommand{}
}
