package main

import (
	"maze_gen/maze"
	"testing"
)

func TestMazeStartAtMaximum(t *testing.T) {
	width := 20
	height := 20

	startX := width - 1
	startY := height - 1

	start := maze.Position{
		X: startX,
		Y: startY,
	}

	m := maze.New(width, height, start)

	cell := m.Cells[startX][startY]

	if !cell.Visited {
		t.Errorf("cell not market visited: cell [%d, %d] w/h: %d%d", startX, startY, width, height)
	}

}
