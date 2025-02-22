// Package utils [general utility functions for the db1000n app]
package utils

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

// PanicHandler just stub it in the beginning of every major module invocation to prevent single module failure from crashing the whole app
func PanicHandler(logger *zap.Logger) {
	if err := recover(); err != nil {
		logger.Error("caught panic, recovering", zap.Any("err", err))
	}
}

// GetEnvStringDefault returns environment variable or default value if no env varible is present
func GetEnvStringDefault(key, defaultValue string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	return value
}

// GetEnvIntDefault returns environment variable or default value if no env varible is present
func GetEnvIntDefault(key string, defaultValue int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return v
}

// GetEnvBoolDefault returns environment variable or default value if no env varible is present
func GetEnvBoolDefault(key string, defaultValue bool) bool {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	v, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}

	return v
}

// GetEnvDurationDefault returns environment variable or default value if no env varible is present
func GetEnvDurationDefault(key string, defaultValue time.Duration) time.Duration {
	value, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	v, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}

	return v
}

func NonNilOrDefault[T any](v *T, dflt T) T {
	if v != nil {
		return *v
	}

	return dflt
}

// Decode is an alias to a mapstructure.NewDecoder({Squash: true}).Decode()
// with WeaklyTypedInput set to true and MatchFunc that only compares aplhanumeric sequence in field names
func Decode(input any, output any) error {
	filter := func(r rune) rune {
		if ('a' <= r && r <= 'z') ||
			('A' <= r && r <= 'Z') ||
			('0' <= r && r <= '9') {
			return r
		}

		return -1
	}

	matchName := func(lhs, rhs string) bool {
		return strings.EqualFold(strings.Map(filter, lhs), strings.Map(filter, rhs))
	}

	decoderConfig := &mapstructure.DecoderConfig{
		Squash:           true,
		WeaklyTypedInput: true,
		MatchName:        matchName,
		DecodeHook:       mapstructure.StringToTimeDurationHookFunc(),
		Result:           output,
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}

// It could've been more effective if go had yield but this should be close enough
func InfiniteRange[T any](ctx context.Context, input []T) chan T {
	result := make(chan T)

	loop := func() {
		defer close(result)

		for {
			for _, el := range input {
				select {
				case <-ctx.Done():
					return
				case result <- el:
				}
			}
		}
	}
	go loop()

	return result
}

func Unmarshal(input []byte, output any, format string) error {
	switch format {
	case "", "json", "yaml":
		if err := yaml.Unmarshal(input, output); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unknown config format: %v", format)
	}

	return nil
}
