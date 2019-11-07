package main

import "strconv"

type optionalBool struct {
	IsSet bool
	Value bool
}

func (s *optionalBool) IsBoolFlag() bool { return true }

func (s *optionalBool) Set(x string) error {
	b, err := strconv.ParseBool(x)
	if err != nil {
		return err
	}
	s.Value = b
	s.IsSet = true
	return nil
}

func (s *optionalBool) String() string {
	if s == nil {
		return ""
	}
	if !s.IsSet {
		return ""
	} else if s.Value {
		return "true"
	} else {
		return "false"
	}
}
