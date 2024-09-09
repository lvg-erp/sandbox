package alg

type MinStack struct {
	stack [][]int
}

func NewMinStack() *MinStack {
	return &MinStack{stack: make([][]int, 0)}
}

func min(a, b int) int {
	if a < b {
		return a
	}

	return b

}

func (t *MinStack) Push(v int) {
	if len(t.stack) == 0 {
		t.stack = append(t.stack, []int{v, v})
		return
	}

	currentMin := t.stack[len(t.stack)-1][1]

	t.stack = append(t.stack, []int{v, min(v, currentMin)})

}

func (t *MinStack) Pop() {
	t.stack = t.stack[:len(t.stack)-1]
}

func (t *MinStack) Top() int {
	return t.stack[len(t.stack)-1][0]
}

func (t *MinStack) GetMin() int {
	return t.stack[len(t.stack)-1][1]
}
