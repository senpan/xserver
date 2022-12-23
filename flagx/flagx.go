package flagx

import (
	"flag"
	"log"
	"os"
	"strings"
)

var confPath = flag.String("c", "", "config path")
var confPathPrefix = flag.String("p", "", "config path prefix with no trailing backslash")
var version = flag.Bool("v", false, "show version")
var task = flag.String("task", "", "job option, use for command script")
var usr1 = flag.String("usr1", "", "user defined flagx -usr1")
var usr2 = flag.String("usr2", "", "user defined flagx -usr2")
var usr3 = flag.String("usr3", "", "user defined flagx -usr3")
var usr4 = flag.String("usr4", "", "user defined flagx -usr4")
var usr5 = flag.String("usr5", "", "user defined flagx -usr5")

func init() {
	args := os.Args[1:]
	internalArgsKeys := []string{"c", "p", "v", "task", "usr1", "usr2", "usr3", "usr4", "usr5"}
	internalArgs := extractInternalArgs(args, internalArgsKeys)
	// Define an internal flagx set
	// Avoid conflicts with users
	internalFlagSet := *flag.CommandLine
	// parse internal args
	err := internalFlagSet.Parse(internalArgs)
	if err != nil {
		log.Printf("flagx init error: %+v", err)
	}
}

func extractInternalArgs(arguments []string, internalArgsKeys []string) []string {
	var args []string
	isInternalArg := false
	for _, arg := range arguments {
		if arg[0] == '-' {
			isInternalArg = false
		}
		for _, internalArgsKey := range internalArgsKeys {
			if arg == "-"+internalArgsKey ||
				arg == "--"+internalArgsKey ||
				strings.HasPrefix(arg, "-"+internalArgsKey+"=") ||
				strings.HasPrefix(arg, "--"+internalArgsKey+"=") {
				isInternalArg = true
				break
			}
		}
		if isInternalArg {
			args = append(args, arg)
		}
	}
	return args
}

func GetVersion() *bool {
	return version
}

func GetConfig() *string {
	return confPath
}

func SetConfig(path string) {
	confPath = &path
}

func GetConfigPathPrefix() *string {
	return confPathPrefix
}

func SetConfigPathPrefix(pathPrefix string) {
	confPathPrefix = &pathPrefix
}

// GetTask for command script
func GetTask() *string {
	return task
}

func GetUsr1() *string {
	return usr1
}
func GetUsr2() *string {
	return usr2
}
func GetUsr3() *string {
	return usr3
}
func GetUsr4() *string {
	return usr4
}
func GetUsr5() *string {
	return usr5
}
