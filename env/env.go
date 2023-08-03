package env

import (
	"os"
	"strconv"
)

// Getter -
func Getter(key, defaultValue string) string {
	env, ok := os.LookupEnv(key)
	if ok {
		return env
	}
	return defaultValue
}

// GetterInt -
func GetterInt(key string, defaultValue int) int {
	env, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}
	intEnv, err := strconv.Atoi(env)
	if err != nil {
		return defaultValue
	}

	return intEnv
}

// GetterBool -
func GetterBool(key string, defaultValue bool) bool {
	env, ok := os.LookupEnv(key)
	if ok {
		res, err := strconv.ParseBool(env)
		if err == nil {
			return res
		}
	}
	return defaultValue
}
