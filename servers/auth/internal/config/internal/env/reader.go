package env

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var (
	ErrRequired   = errors.New("is required")
	ErrEmpty      = errors.New("cannot be empty")
	ErrNotInteger = errors.New("must be an integer")
	ErrOutOfRange = errors.New("out of range")
)

type Reader struct {
	errs []error
}

func (r *Reader) lookup(name string) (string, bool) {
	value, ok := os.LookupEnv(name)

	if !ok {
		r.errs = append(r.errs, fmt.Errorf("%s %w", name, ErrRequired))
		return "", false
	}

	value = strings.TrimSpace(value)

	if value == "" {
		r.errs = append(r.errs, fmt.Errorf("%s %w", name, ErrEmpty))
		return "", false
	}

	return value, true
}

func (r *Reader) String(name string) string {
	value, _ := r.lookup(name)
	return value
}

func (r *Reader) IntRange(name string, min, max int) int {
	v, ok := r.parseInt(name)
	if !ok {
		return 0
	}

	if v < min || v > max {
		r.errs = append(r.errs, fmt.Errorf("%s %w [%d, %d], got %d", name, ErrOutOfRange, min, max, v))
		return 0
	}

	return v
}

func (r *Reader) parseInt(name string) (int, bool) {
	value, ok := r.lookup(name)
	if !ok {
		return 0, false
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		r.errs = append(r.errs, fmt.Errorf("%s %w, got %q", name, ErrNotInteger, value))
		return 0, false
	}

	return v, true
}

func (r *Reader) Err() error {
	return errors.Join(r.errs...)
}
