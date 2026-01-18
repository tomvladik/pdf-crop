package cli

import (
	"fmt"
	"strconv"
)

// RequireValue returns the next argument after a flag or an error if missing.
func RequireValue(argv []string, i int, flag string) (string, int, error) {
	if i+1 >= len(argv) {
		return "", i, fmt.Errorf("missing value for %s", flag)
	}
	return argv[i+1], i + 1, nil
}

// RequireValues returns the next n arguments after a flag or an error if insufficient.
func RequireValues(argv []string, i, n int, flag string) ([]string, int, error) {
	if i+n >= len(argv) {
		return nil, i, fmt.Errorf("%s requires %d arguments", flag, n)
	}
	return argv[i+1 : i+1+n], i + n, nil
}

// ParseInt parses an int value and annotates errors with the flag name.
func ParseInt(val string, flag string) (int, error) {
	v, err := strconv.Atoi(val)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", flag, err)
	}
	return v, nil
}

// ParseFloat parses a float value and annotates errors with the flag name.
func ParseFloat(val string, flag string) (float64, error) {
	v, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", flag, err)
	}
	return v, nil
}
