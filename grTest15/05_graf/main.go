package main

import (
	"container/heap"
	"fmt"
	"math"
	"os"
)

// Edge представляет ребро графа
type Edge struct {
	to     int
	weight float64
}

// Graph представляет граф
type Graph struct {
	adjacency map[int][]Edge
}

// NewGraph создает новый граф
func NewGraph() *Graph {
	return &Graph{
		adjacency: make(map[int][]Edge),
	}
}

// AddEdge добавляет ребро в граф
func (g *Graph) AddEdge(from, to int, weight float64) {
	g.adjacency[from] = append(g.adjacency[from], Edge{to: to, weight: weight})
}

// AddBidirectionalEdge — добавляет рёбра в оба направления
func (g *Graph) AddBidirectionalEdge(from, to int, weight float64) {
	g.AddEdge(from, to, weight)
	g.AddEdge(to, from, weight)
}

// Dijkstra ищет кратчайший путь от start до end
func (g *Graph) Dijkstra(start, end int) ([]int, float64) {
	dist := make(map[int]float64)
	prev := make(map[int]int)
	for node := range g.adjacency {
		dist[node] = math.Inf(1)
	}
	dist[start] = 0

	pq := &PriorityQueue{}
	heap.Init(pq)
	heap.Push(pq, &Item{node: start, priority: 0})

	for pq.Len() > 0 {
		current := heap.Pop(pq).(*Item)
		u := current.node

		if u == end {
			break
		}

		for _, e := range g.adjacency[u] {
			alt := dist[u] + e.weight
			if alt < dist[e.to] {
				dist[e.to] = alt
				prev[e.to] = u
				heap.Push(pq, &Item{node: e.to, priority: alt})
			}
		}
	}

	// Восстановление маршрута
	var path []int
	u := end
	if _, ok := dist[end]; !ok || math.IsInf(dist[end], 1) {
		return nil, math.Inf(1)
	}
	for u != start {
		path = append([]int{u}, path...)
		u, _ = prev[u]
	}
	path = append([]int{start}, path...)
	return path, dist[end]
}

// PriorityQueue реализует очередь с приоритетом
type Item struct {
	node     int
	priority float64
}

// PriorityQueue реализация `heap.Interface`
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].priority < pq[j].priority
}
func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}
func (pq *PriorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*Item))
}
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func main() {
	g := NewGraph()

	// Добавление рёбер
	g.AddBidirectionalEdge(1, 2, 2)
	g.AddBidirectionalEdge(2, 3, 2)
	g.AddBidirectionalEdge(1, 3, 5)
	g.AddBidirectionalEdge(3, 4, 1)
	g.AddBidirectionalEdge(2, 4, 4)

	// Генерация графа в формат DOT
	generateDot(g, "graph_full.dot", nil)

	start := 1
	waypoints := []int{2, 4}
	end := 4

	points := append(waypoints, end)

	var totalPath []int
	totalDistance := 0.0
	currentStart := start

	for _, point := range points {
		path, dist := g.Dijkstra(currentStart, point)
		if path == nil {
			fmt.Printf("Путь из %d в %d не найден\n", currentStart, point)
			return
		}
		// Объединение маршрутов
		if len(totalPath) == 0 {
			totalPath = path
		} else {
			totalPath = append(totalPath, path[1:]...)
		}
		totalDistance += dist
		currentStart = point
	}

	fmt.Printf("Общий маршрут: %v\n", totalPath)
	fmt.Printf("Общая длина: %.2f\n", totalDistance)

	// Генерация графа с маршрутом выделенным
	generateDot(g, "graph_with_route.dot", totalPath)
}

// Функция для генерации файла .dot
func generateDot(g *Graph, filename string, route []int) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	fmt.Fprintln(f, "graph G {")
	// Создаем карту для быстрого поиска рёбер маршрута
	routeEdges := make(map[[2]int]bool)
	if route != nil && len(route) > 1 {
		for i := 0; i < len(route)-1; i++ {
			a, b := route[i], route[i+1]
			routeEdges[[2]int{a, b}] = true
			routeEdges[[2]int{b, a}] = true // двунаправлено
		}
	}

	// вывод всех рёбер
	for from, edges := range g.adjacency {
		for _, e := range edges {
			// если ребро в маршруте, выделим его
			key1 := [2]int{from, e.to}
			key2 := [2]int{e.to, from}
			if route != nil && (routeEdges[key1] || routeEdges[key2]) {
				// выделенное ребро — красным
				fmt.Fprintf(f, "    %d -- %d [color=red, penwidth=2, label=\"%.0f\"];\n", from, e.to, e.weight)
			} else {
				// обычное
				fmt.Fprintf(f, "    %d -- %d [label=\"%.0f\"];\n", from, e.to, e.weight)
			}
		}
	}
	fmt.Fprintln(f, "}")
	fmt.Printf("Файл %s создан.\n", filename)
}
