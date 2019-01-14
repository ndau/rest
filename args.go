package rest

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Config maintains a map of configuration values. They can be read from
// the environment and/or the command line (command line overrides environment).
// The names of configuration items are strings containing letters, numbers,
// underscores, and hyphens. They can be any of int, string, []string, duration,
// or bool.
// Although there are lots of configuration packages out there, there didn't
// seem to be one that was both simple and would allow the standard service
// to define part of the configuration, and client packages to define the
// rest.
type Config map[string]ConfigItem

// ConfigItem is one element of the Config map
type ConfigItem struct {
	Name    string
	Type    string
	Value   interface{}
	Default interface{}
}

func clean(s string) string {
	return strings.TrimSpace(strings.ToUpper(strings.Replace(s, "-", "_", -1)))
}

// AddInt adds a config element that is an integer with its default.
func (cf *Config) AddInt(name string, def int) {
	(*cf)[clean(name)] = ConfigItem{Name: name, Type: "int", Default: def}
}

// AddString adds a config element that is a string with its default.
func (cf *Config) AddString(name string, def string) {
	(*cf)[clean(name)] = ConfigItem{Name: name, Type: "string", Default: def}
}

// AddStringArray adds a config element that is an array of strings with an arbitrary
// list of default values
func (cf *Config) AddStringArray(name string, defaults ...string) {
	(*cf)[clean(name)] = ConfigItem{Name: name, Type: "[]string", Default: defaults}
}

// AddFlag adds a config element that is a boolean flag with a default value.
func (cf *Config) AddFlag(name string, def bool) {
	(*cf)[clean(name)] = ConfigItem{Name: name, Type: "bool", Default: def}
}

// AddDuration adds a config element that is a duration with a default value.
// The duration is specified as a string and is returned as a time.Duration.
func (cf *Config) AddDuration(name string, def string) {
	(*cf)[clean(name)] = ConfigItem{Name: name, Type: "duration", Default: def}
}

// AddRequiredInt adds a config element that is an integer with no default value
// (it must be specified or the server will fail to start).
func (cf *Config) AddRequiredInt(name string) {
	(*cf)[clean(name)] = ConfigItem{Name: name, Type: "int", Default: nil}
}

// AddRequiredString adds a config element that is a string with no default value
// (it must be specified or the server will fail to start).
func (cf *Config) AddRequiredString(name string) {
	(*cf)[clean(name)] = ConfigItem{Name: name, Type: "string", Default: nil}
}

// AddRequiredFlag adds a config element that is a boolean with no default value
// (it must be specified or the server will fail to start).
func (cf *Config) AddRequiredFlag(name string) {
	(*cf)[clean(name)] = ConfigItem{Name: name, Type: "bool", Default: nil}
}

// SetDefault allows setting a default value for a name after it has been
// created. It does no error checking and if the default type doesn't agree
// with the config item's type, weird things can happen.
func (cf *Config) SetDefault(name string, def interface{}) {
	name = clean(name)
	ag := (*cf)[name]
	ag.Default = def
	(*cf)[name] = ag
}

// parseValue tries to parse a value from a string
// according to the type hint. It doesn't do errors.
func parseValue(s string, typ string) interface{} {
	var v interface{}
	switch typ {
	case "int":
		v, _ = strconv.Atoi(s)
	case "string":
		v = s
	case "[]string":
		v = strings.Split(s, ",")
	case "bool":
		// for flags, simply specifying the name means "true"
		if s == "" || strings.HasPrefix(strings.ToLower(s), "t") {
			v = true
		} else {
			v = false
		}
	case "duration":
		if s == "" {
			return time.Duration(0)
		}
		d, err := time.ParseDuration(s)
		if err != nil {
			return time.Duration(0)
		}
		return d
	}
	return v
}

// ParseCmdLine parses the command line and stores the values it finds
// into the Config. Command line flags are expected to start with
// either 1 or 2 leading hyphens (no short codes are supported), and are
// case-insensitive. Hyphens and underscores (after the leading ones) are
// equivalent. Values *must* be specified with an equals sign and be
// part of the same argument, so `--foo=bar` is good, `--foo bar` is bad.
func (cf *Config) ParseCmdLine() {
	argp := regexp.MustCompile(`^--?([A-Za-z0-9_-]+)(?:=(.+))?$`)
	for ix := 1; ix < len(os.Args); ix++ {
		if argp.MatchString(os.Args[ix]) {
			m := argp.FindStringSubmatch(os.Args[ix])
			name := clean(m[1])
			value := m[2]
			ag, ok := (*cf)[name]
			if ok {
				ag.Value = parseValue(value, ag.Type)
				(*cf)[name] = ag
			} else {
				fmt.Println("unrecognized command line argument: ", os.Args[ix])
				os.Exit(1)
			}
		}
	}
}

// ParseEnv reads the config and looks for environment variables that match,
// parsing their values appropriately and overwriting existing configs
func (cf *Config) ParseEnv() {
	for name := range *cf {
		ag := (*cf)[name]
		value := os.Getenv(clean(name))
		if value != "" {
			ag.Value = parseValue(value, ag.Type)
			(*cf)[name] = ag
		}
	}
}

// Check walks the config and looks for required config items that were
// not specified; if it finds any, it logs it and kills the server.
func (cf *Config) Check() {
	for name := range *cf {
		ag := (*cf)[name]
		if ag.Value == nil && ag.Default == nil {
			fmt.Printf("required %s parameter %s was not found", ag.Type, ag.Name)
			os.Exit(1)
		}
	}
}

// Get is a generic Get that returns an interface and a flag if it was
// found to be a valid config variable.
func (cf *Config) Get(name string) (interface{}, bool) {
	ag, ok := (*cf)[clean(name)]
	if !ok {
		return nil, false
	}
	if ag.Value != nil {
		return ag.Value, true
	}
	return ag.Default, true
}

// GetInt retrieves an integer from the config, or 0 if not found.
func (cf *Config) GetInt(name string) int {
	v, ok := cf.Get(name)
	if !ok {
		return 0
	}
	return v.(int)
}

// GetString retrieves a string from the config, or "" if not found.
func (cf *Config) GetString(name string) string {
	v, ok := cf.Get(name)
	if !ok {
		return ""
	}
	return v.(string)
}

// GetDuration retrieves a duration from the config as a time.Duration.
func (cf *Config) GetDuration(name string) time.Duration {
	v, ok := cf.Get(name)
	if !ok {
		return time.Duration(0)
	}
	return v.(time.Duration)
}

// GetFlag retrieves a boolean from the config, or false if not found.
func (cf *Config) GetFlag(name string) bool {
	v, ok := cf.Get(name)
	if !ok {
		return false
	}
	return v.(bool)
}

// GetStringArray retrieves a string array from the config, or []string{} if not found.
func (cf *Config) GetStringArray(name string) []string {
	v, ok := cf.Get(name)
	if !ok {
		return []string{}
	}
	return v.([]string)
}

// Load fetches config values from the environment, and then from the command
// line, then checks to see if any required variables were missing.
// Envvars are higher priority than default values, and cmd line is higher
// priority than envvars.
func (cf *Config) Load() {
	cf.ParseEnv()
	cf.ParseCmdLine()
	cf.Check()
}

// NewConfig constructs an empty config
func NewConfig() *Config {
	a := &Config{}
	return a
}
