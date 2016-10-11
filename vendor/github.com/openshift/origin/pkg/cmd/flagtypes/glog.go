package flagtypes

import (
	"flag"

	"github.com/golang/glog"
	"github.com/spf13/pflag"
)

// GLog binds the log flags from the default Google "flag" package into a pflag.FlagSet.
func GLog(flags *pflag.FlagSet) {
	from := flag.CommandLine
	if flag := from.Lookup("v"); flag != nil {
		level := flag.Value.(*glog.Level)
		levelPtr := (*int32)(level)
		flags.Int32Var(levelPtr, "loglevel", 0, "Set the level of log output (0-10)")
	}
	if flag := from.Lookup("vmodule"); flag != nil {
		value := flag.Value
		flags.Var(pflagValue{value}, "logspec", "Set per module logging with file|pattern=LEVEL,...")
	}
}

type pflagValue struct {
	flag.Value
}

func (pflagValue) Type() string {
	return "string"
}
