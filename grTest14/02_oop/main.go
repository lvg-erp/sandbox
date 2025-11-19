package main

import "math"

type SimpleCircle struct {
	r float64
}

type SimpleRectangle struct {
	r float64
	b float64
}

type Shape interface {
	Area() float64
}

func (sc *SimpleCircle) Area() float64 {
	return math.Pi * sc.r * sc.r
}

func (sr *SimpleRectangle) Area() float64 {
	return sr.r * sr.b
}

func main() {

}
