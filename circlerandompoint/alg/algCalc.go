package alg

import (
	"golang.org/x/exp/rand"
	"math"
	"time"
)

type Solution struct {
	rad, xc, yc float64
}

func NewSolution(rad, x_center, y_center float64) *Solution {
	rand.Seed(uint64(time.Now().UnixNano()))

	return &Solution{
		rad: rad,
		xc:  x_center,
		yc:  y_center,
	}
}

func (s *Solution) RandPoint() []float64 {
	x0 := s.xc - s.rad
	y0 := s.yc - s.rad

	for {
		xg := x0 + rand.Float64()*s.rad*2
		yg := y0 + rand.Float64()*s.rad*2

		if math.Sqrt(math.Pow(xg-s.xc, 2)+math.Pow(yg-s.yc, 2)) <= s.rad {
			return []float64{xg, yg}
		}

	}
}
