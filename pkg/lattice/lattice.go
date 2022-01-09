package lattice

import (
	"fmt"

	"aaronstgeorge.com/self-guided-cs-1620/pkg/utils"
)

type Lattice[T any] interface {
	// Stringer is used for change comparison in the transfer functions
	fmt.Stringer
	Meet(l T) T
}

var _ Lattice[UnionMeetSetLattice] = UnionMeetSetLattice{}

type UnionMeetSetLattice struct {
	utils.Set
}

func (s UnionMeetSetLattice) Meet(l UnionMeetSetLattice) UnionMeetSetLattice {
	return UnionMeetSetLattice{utils.Union(s.Set, l.Set)}
}

var _ Lattice[IntersetMeetSetLattice] = IntersetMeetSetLattice{}

type IntersetMeetSetLattice struct {
	utils.Set
}

func (s IntersetMeetSetLattice) Meet(l IntersetMeetSetLattice) IntersetMeetSetLattice {
	return IntersetMeetSetLattice{utils.Intersect(s.Set, l.Set)}
}
