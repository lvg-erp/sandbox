package alg

import (
	"golang.org/x/exp/rand"
	"time"
)

type ListNode struct {
	Val  int
	Next *ListNode
}

type Solution struct {
	rangeVals []int
}

func NestedListToLinkedList(nestedList [][][]int) *ListNode {
	dummyHead := &ListNode{} // Dummy head for easier list management
	current := dummyHead

	for _, sublist := range nestedList {
		for _, innerList := range sublist {
			for _, item := range innerList {
				current.Next = &ListNode{Val: item} // Create a new ListNode for each item
				current = current.Next              // Move to the newly created node
			}
		}
	}
	return dummyHead.Next // Return the next of dummy head, which is the actual head of the linked list
}

//func NewListNode(val int) *ListNode {
//	return &ListNode{
//		Val: val,
//	}
//}

func NewSolution(head *ListNode) *Solution {
	var rangeVals []int
	for head != nil {
		rangeVals = append(rangeVals, head.Val)
		head = head.Next
	}

	rand.Seed(uint64(time.Now().UnixNano()))

	return &Solution{rangeVals: rangeVals}
}

func (s *Solution) GetRandom() int {
	pick := rand.Intn(len(s.rangeVals))
	return s.rangeVals[pick]
}
