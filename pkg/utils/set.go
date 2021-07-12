package utils

import (
	"sort"
	"strings"
)

type Set map[string]struct{}

func (s Set) String() string {
	if len(s) == 0 {
		return "∅"
	}
	var items []string
	for k := range s {
		items = append(items, k)
	}
	sort.Strings(items)
	return strings.Join(items, ", ")
}

func (s Set) Add(items ...string) {
	for _, item := range items {
		s[item] = struct{}{}
	}
}

func (s Set) Contains(item string) bool {
	_, ok := s[item]
	return ok
}

func (s Set) Remove(item string) {
	delete(s, item)
}

func NewSet(items ...string) Set {
	out := make(Set)
	out.Add(items...)
	return out
}

func Sub(rhs, lhs Set) Set {
	out := make(Set)
	for item := range rhs {
		out.Add(item)
	}
	for item := range lhs {
		out.Remove(item)
	}
	return out
}

func Union(rhs, lhs Set) Set {
	out := make(Set)
	for item := range rhs {
		out.Add(item)
	}
	for item := range lhs {
		out.Add(item)
	}
	return out
}

func Intersect(rhs, lhs Set) Set {
	out := make(Set)
	for item := range rhs {
		if lhs.Contains(item) {
			out.Add(item)
		}
	}
	return out
}
