package maze

type Position struct {
	X int
	Y int
}

type Route []Position

func (s *Route) Push(p Position) {
	*s = append(*s, p)
}

func (s *Route) Pop() Position {
	l := len(*s)
	p := (*s)[l-1]
	*s = (*s)[:l-1]

	return p

}

// ячейка — это отдельная часть лабиринта с маршрутом на север, запад и флагом посещения
// «маршрут» на север, обозначающий отсутствие стены, используется вместо wall=true как
// тип bool по умолчанию — false, поэтому по умолчанию маршрут из этой ячейки отсутствует.
// Это избавляет от необходимости создавать ячейки с истинными значениями.
type Cell struct {
	NorthRoute bool
	WestRoute  bool
	Visited    bool
}

// лабиринт представляет весь лабиринт
type Maze struct {
	Cells [][]Cell
}

func New(width, height int, start Position) Maze {
	route := make(Route, 0)

	var m Maze

	m.Cells = make([][]Cell, width)
	for i := range m.Cells {
		m.Cells[i] = make([]Cell, height)
	}
	m.generate(start, &route)

	return m
}

func (m *Maze) generate(p Position, r *Route) {
	r.Push(p)
	c := &m.Cells[p.Y][p.X]
	c.Visited = true

}
