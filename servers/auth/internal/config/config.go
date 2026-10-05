package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port int
}

type envReader struct {
	errs []error
}

func (r *envReader) lookup(name string) (string, bool) {
	value, ok := os.LookupEnv(name)

	if !ok {
		r.errs = append(r.errs, fmt.Errorf("%s é obrigatorio", name))
		return "", false
	}

	value = strings.TrimSpace(value)

	if value == "" {
		r.errs = append(r.errs, fmt.Errorf("%s não pode ser vazia", name))
		return "", false
	}

	return value, true
}

func requiredField(name string, r *envReader) string {
	value, ok := os.LookupEnv(name)

	if !ok {
		r.errs = append(r.errs, fmt.Errorf("%s é obrigatorio", name))
		return ""
	}

	return value
}

func (r *envReader) string(name string) string {
	value, _ := r.lookup(name)
	return value
}

func (r *envReader) intRange(name string, min, max int) int {
	v, ok := r.parseInt(name)
	if !ok {
		return 0
	}

	if v < min || v > max {
		r.errs = append(r.errs, fmt.Errorf("%s deve estar entre %d e %d, recebido %d", name, min, max, v))
		return 0
	}

	return v
}

func (r *envReader) parseInt(name string) (int, bool) {
	value, ok := r.lookup(name)
	if !ok {
		return 0, false
	}

	v, err := strconv.Atoi(value)
	if err != nil {
		r.errs = append(r.errs, fmt.Errorf("%s deve ser um número inteiro, recebido %q", name, value))
		return 0, false
	}

	return v, true
}

func (r envReader) err() error {
	return errors.Join(r.errs...)
}

func Load() (Config, error) {

	r := envReader{}
	c := Config{
		Port: r.intRange("PORT", 1, 65535),
	}

	if err := r.err(); err != nil {
		return Config{}, err
	}

	return c, nil
}
