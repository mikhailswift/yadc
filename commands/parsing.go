package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-shellwords"
)

const (
	getNumArgs        = 1
	setNumArgs        = 2
	setWithTTLNumArgs = 3
	unsetNumArgs      = 0
	getTTLNumArgs     = 1
	setTTLNumArgs     = 2
	unsetTTLNumArgs   = 1
	unsetAllNumArgs   = 0
)

var (
	commandParsers = map[string]parseFn{
		"GET":      parseGet,
		"SET":      parseSet,
		"UNSET":    parseUnset,
		"GETTTL":   parseGetTTL,
		"SETTTL":   parseSetTTL,
		"UNSETTTL": parseUnsetTTL,
		"UNSETALL": parseUnsetAll,
	}

	parser = shellwords.NewParser()
)

type parseFn func(cmdStr string, args ...string) (Command, error)

//ErrCmdNotFound is the error that is returned when ParseCommand is handed a string it cannot recognize as a command.
type ErrCmdNotFound string

func (e ErrCmdNotFound) Error() string {
	return fmt.Sprintf("Couldn't find command %v", string(e))
}

//ErrInvalidArgs is the error that is returned when the command gets an unexpected number of arguments
type ErrInvalidArgs struct {
	ActualLen   int
	ExpectedLen int
	Cmd         string
	Args        []string
}

func (e ErrInvalidArgs) Error() string {
	return fmt.Sprintf("Invalid arguments for %v command. Expected %v args but got %v args", e.Cmd, e.ExpectedLen, e.ActualLen)
}

//ErrInvalidArg is the error returned when an argument couldn't properly be parsed for a command
type ErrInvalidArg struct {
	Arg      string
	ArgValue string
	Cmd      string
	ParseErr error
}

func (e ErrInvalidArg) Error() string {
	return fmt.Sprintf("Invalid argument for %v command.  Couldn't parse %v as the %v argument.", e.Cmd, e.ArgValue, e.Arg)
}

//ParseCommand will parse a string and turn it into a Command.
//Example: `GET "Key"` will return a getCommand that will get the key "Key" from the cache.
func ParseCommand(cmdStr string) (Command, error) {
	cmd, args, err := splitCmdString(cmdStr)
	if err != nil {
		return nil, err
	}

	cmdParserFn, ok := commandParsers[cmd]
	if !ok {
		return nil, ErrCmdNotFound(cmd)
	}

	return cmdParserFn(cmd, args...)
}

func splitCmdString(cmdStr string) (cmd string, args []string, err error) {
	splitCmdString, err := parser.Parse(cmdStr)
	if err != nil {
		return cmd, args, err
	}

	if len(splitCmdString) == 0 {
		return cmd, args, ErrCmdNotFound(cmdStr)
	}

	cmd = strings.ToUpper(splitCmdString[0])
	args = splitCmdString[1:]
	return cmd, args, err
}

func checkNumArgs(cmd string, args []string, expectedArgs int) error {
	numArgs := len(args)
	if numArgs != expectedArgs {
		return ErrInvalidArgs{
			Cmd:         cmd,
			ActualLen:   numArgs,
			ExpectedLen: expectedArgs,
			Args:        args,
		}
	}

	return nil
}

func parseGet(cmdStr string, args ...string) (Command, error) {
	if err := checkNumArgs(cmdStr, args, getNumArgs); err != nil {
		return nil, err
	}

	return NewGetCommand(args[0]), nil
}

func parseSet(cmdStr string, args ...string) (Command, error) {
	argsLen := len(args)
	if argsLen != setWithTTLNumArgs && argsLen != setNumArgs {
		return nil, ErrInvalidArgs{
			ActualLen:   argsLen,
			ExpectedLen: 2,
			Cmd:         cmdStr,
			Args:        args,
		}
	}

	var ttl time.Duration
	if argsLen == 3 {
		secs, err := strconv.Atoi(args[2])
		if err != nil {
			return nil, ErrInvalidArg{
				Arg:      "TTL",
				ArgValue: args[2],
				Cmd:      cmdStr,
				ParseErr: err,
			}
		}

		ttl = time.Duration(secs) * time.Second
	}

	return NewSetCommand(args[0], args[1], ttl), nil
}

func parseUnset(cmdStr string, args ...string) (Command, error) {
	if err := checkNumArgs(cmdStr, args, unsetNumArgs); err != nil {
		return nil, err
	}

	return NewUnsetCommand(args[0]), nil
}

func parseGetTTL(cmdStr string, args ...string) (Command, error) {
	if err := checkNumArgs(cmdStr, args, getTTLNumArgs); err != nil {
		return nil, err
	}

	return NewGetTTLCommand(args[0]), nil
}

func parseSetTTL(cmdStr string, args ...string) (Command, error) {
	if err := checkNumArgs(cmdStr, args, setTTLNumArgs); err != nil {
		return nil, err
	}

	secs, err := strconv.Atoi(args[1])
	if err != nil {
		return nil, ErrInvalidArg{
			Arg:      "TTL",
			ArgValue: args[2],
			Cmd:      cmdStr,
			ParseErr: err,
		}
	}

	ttl := time.Duration(secs) * time.Second

	return NewSetTTLCommand(args[0], ttl), nil
}

func parseUnsetTTL(cmdStr string, args ...string) (Command, error) {
	if err := checkNumArgs(cmdStr, args, unsetTTLNumArgs); err != nil {
		return nil, err
	}

	return NewUnsetTTLCommand(args[0]), nil
}

func parseUnsetAll(cmdStr string, args ...string) (Command, error) {
	if err := checkNumArgs(cmdStr, args, unsetTTLNumArgs); err != nil {
		return nil, err
	}

	return NewUnsetAllCommand(), nil
}
