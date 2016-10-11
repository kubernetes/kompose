package util

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/util/sets"
)

func EnvInt(key string, defaultValue int32, minValue int32) int32 {
	value, err := strconv.ParseInt(Env(key, fmt.Sprintf("%d", defaultValue)), 10, 32)
	if err != nil || int32(value) < minValue {
		return defaultValue
	}
	return int32(value)
}

// Env returns an environment variable or a default value if not specified.
func Env(key string, defaultValue string) string {
	val := os.Getenv(key)
	if len(val) == 0 {
		return defaultValue
	}
	return val
}

// GetEnv returns an environment value if specified
func GetEnv(key string) (string, bool) {
	val := os.Getenv(key)
	if len(val) == 0 {
		return "", false
	}
	return val, true
}

type Environment map[string]string

var argumentEnvironment = regexp.MustCompile("(?ms)^(.+)\\=(.*)$")
var validArgumentEnvironment = regexp.MustCompile("(?ms)^(\\w+)\\=(.*)$")

func IsEnvironmentArgument(s string) bool {
	return argumentEnvironment.MatchString(s)
}

func IsValidEnvironmentArgument(s string) bool {
	return validArgumentEnvironment.MatchString(s)
}

func SplitEnvironmentFromResources(args []string) (resources, envArgs []string, ok bool) {
	first := true
	for _, s := range args {
		// this method also has to understand env removal syntax, i.e. KEY-
		isEnv := IsEnvironmentArgument(s) || strings.HasSuffix(s, "-")
		switch {
		case first && isEnv:
			first = false
			fallthrough
		case !first && isEnv:
			envArgs = append(envArgs, s)
		case first && !isEnv:
			resources = append(resources, s)
		case !first && !isEnv:
			return nil, nil, false
		}
	}
	return resources, envArgs, true
}

func ParseEnvironmentArguments(s []string) (Environment, []string, []error) {
	errs := []error{}
	duplicates := []string{}
	env := make(Environment)
	for _, s := range s {
		switch matches := validArgumentEnvironment.FindStringSubmatch(s); len(matches) {
		case 3:
			k, v := matches[1], matches[2]
			if exist, ok := env[k]; ok {
				duplicates = append(duplicates, fmt.Sprintf("%s=%s", k, exist))
			}
			env[k] = v
		default:
			errs = append(errs, fmt.Errorf("environment variables must be of the form key=value: %s", s))
		}
	}
	return env, duplicates, errs
}

// ParseEnv parses the list of environment variables into kubernetes EnvVar
func ParseEnv(spec []string, defaultReader io.Reader) ([]kapi.EnvVar, []string, error) {
	env := []kapi.EnvVar{}
	exists := sets.NewString()
	var remove []string
	for _, envSpec := range spec {
		switch {
		case !IsValidEnvironmentArgument(envSpec) && !strings.HasSuffix(envSpec, "-"):
			return nil, nil, fmt.Errorf("environment variables must be of the form key=value and can only contain letters, numbers, and underscores")
		case envSpec == "-":
			if defaultReader == nil {
				return nil, nil, fmt.Errorf("when '-' is used, STDIN must be open")
			}
			fileEnv, err := readEnv(defaultReader)
			if err != nil {
				return nil, nil, err
			}
			env = append(env, fileEnv...)
		case strings.Index(envSpec, "=") != -1:
			parts := strings.SplitN(envSpec, "=", 2)
			if len(parts) != 2 {
				return nil, nil, fmt.Errorf("invalid environment variable: %v", envSpec)
			}
			exists.Insert(parts[0])
			env = append(env, kapi.EnvVar{
				Name:  parts[0],
				Value: parts[1],
			})
		case strings.HasSuffix(envSpec, "-"):
			remove = append(remove, envSpec[:len(envSpec)-1])
		default:
			return nil, nil, fmt.Errorf("unknown environment variable: %v", envSpec)
		}
	}
	for _, removeLabel := range remove {
		if _, found := exists[removeLabel]; found {
			return nil, nil, fmt.Errorf("can not both modify and remove an environment variable in the same command")
		}
	}
	return env, remove, nil
}

func readEnv(r io.Reader) ([]kapi.EnvVar, error) {
	env := []kapi.EnvVar{}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		envSpec := scanner.Text()
		if pos := strings.Index(envSpec, "#"); pos != -1 {
			envSpec = envSpec[:pos]
		}
		if strings.Index(envSpec, "=") != -1 {
			parts := strings.SplitN(envSpec, "=", 2)
			if len(parts) != 2 {
				return nil, fmt.Errorf("invalid environment variable: %v", envSpec)
			}
			env = append(env, kapi.EnvVar{
				Name:  parts[0],
				Value: parts[1],
			})
		}
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, err
	}
	return env, nil
}
