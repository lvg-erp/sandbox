package main

import (
	"flag"
	"io"
	"math/rand"
	"maze_gen/graphics"
	"maze_gen/maze"
	"os"
	"time"
)

var width = flag.Int("width", 20, "width of the maze")
var height = flag.Int("height", 20, "height of the maze")

func main() {
	flag.Parse()
	generateMaze(os.Stdout)
}

func generateMaze(w io.Writer) {
	//rand.Seed(time.Now().UnixNano())
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)
	start := maze.Position{
		X: rng.Intn(*width - 1),
		Y: rng.Intn(*height - 1),
	}

	m := maze.New(*width, *height, start)
	graphics.Render(w, &m, start)
}
