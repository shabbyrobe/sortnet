package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var sizePattern = regexp.MustCompile(`` +
	`^` +
	`(?P<size>[0-9]+)` +
	`(?:-(?P<to>[0-9]+))?` +
	`$`,
)

const (
	sizeSize = 1
	sizeTo   = 2
)

type sizeSpec struct {
	orig  string
	items []int
}

func (sz sizeSpec) String() string {
	return sz.orig
}

func (sz *sizeSpec) Set(s string) error {
	sz.orig = s
	sz.items = sz.items[:0]

	sizeSet := make(map[int]bool)

	for _, sizePart := range strings.Split(s, ",") {
		sizeMatch := sizePattern.FindStringSubmatch(sizePart)
		if len(sizeMatch) == 0 {
			return fmt.Errorf("invalid size %q", sizePart)
		}

		from, err := strconv.ParseInt(sizeMatch[sizeSize], 10, 0)
		if err != nil {
			return fmt.Errorf("size was not numeric in %q", sizePart)
		}
		to := from
		if sizeMatch[sizeTo] != "" {
			to, err = strconv.ParseInt(sizeMatch[sizeTo], 10, 0)
			if err != nil {
				return fmt.Errorf("size range end was not numeric in %q", sizePart)
			}
		}
		if to < from {
			return fmt.Errorf("size range end was before start in %q", sizePart)
		}

		for i := from; i <= to; i++ {
			if i <= 0 {
				return fmt.Errorf("sort size must be >= 1, found %d", i)
			}
			sizeSet[int(i)] = true
		}
	}

	for k := range sizeSet {
		sz.items = append(sz.items, k)
	}

	sort.Ints(sz.items)

	if len(sz.items) == 0 {
		return fmt.Errorf("input %q contained no sizes", s)
	}

	return nil
}
