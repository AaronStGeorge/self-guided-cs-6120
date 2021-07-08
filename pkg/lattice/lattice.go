package lattice

import (
	"fmt"

	"oooga.ooo/cs-1620/pkg/utils"
)

type Lattice interface {
	// Stringer is used for change comparison in the transfer functions
	fmt.Stringer
	// Oh how I pine for generics
	Meet(l Lattice) Lattice
}

var _ Lattice = UnionMeetSetLattice{}

type UnionMeetSetLattice struct {
	utils.Set
}

func (s UnionMeetSetLattice) Meet(l Lattice) Lattice {
	return UnionMeetSetLattice{utils.Union(s.Set, l.(UnionMeetSetLattice).Set)}
}

var _ Lattice = IntersetMeetSetLattice{}

type IntersetMeetSetLattice struct {
	utils.Set
}

func (s IntersetMeetSetLattice) Meet(l Lattice) Lattice {
	return IntersetMeetSetLattice{utils.Intersect(s.Set, l.(IntersetMeetSetLattice).Set)}
}
