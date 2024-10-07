package alg

type MyStack struct {
	q1         []int
	q2         []int
	topElement int
}

func NewStack() MyStack {
	return MyStack{
		q1: []int{},
		q2: []int{},
	}
}

func (s *MyStack) Push(v int) {
	s.q2 = append(s.q2, v)
	s.topElement = v
	for len(s.q1) > 0 {
		s.q2 = append(s.q2, s.q1[0])
		s.q1 = s.q1[1:]
	}

	s.q1, s.q2 = s.q2, s.q1
}

func (s *MyStack) Pop() int {
	res := s.q1[0]
	s.q1 = s.q1[1:]
	if len(s.q1) > 0 {
		s.topElement = s.q1[0]
	}

	return res

}

func (s *MyStack) Top() int {
	return s.topElement
}

func (s *MyStack) Empty() bool {
	return len(s.q2) == 0
}
