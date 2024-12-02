package alg

import (
	"strconv"
)

type ListNode struct {
	Val  int
	Next *ListNode
}

func SwapPairs(head *ListNode) *ListNode {
	for cur := head; cur != nil && cur.Next != nil; cur = cur.Next.Next {
		cur.Val, cur.Next.Val = cur.Next.Val, cur.Val
	}

	return head

}

func NewList(nestedList string) (*ListNode, error) {
	dummyHead := &ListNode{} // Dummy head for easier list management
	current := dummyHead

	for _, item := range nestedList {
		k, err := strconv.Atoi(string(item))
		if err != nil {
			return nil, err
		}
		current.Next = &ListNode{Val: k} // Create a new ListNode for each item
		current = current.Next           // Move to the newly created node

	}
	return dummyHead.Next, nil // Return the next of dummy head, which is the actual head of the linked list
}

type Solution struct {
	rangeVals []int
}

func NewSolution(head *ListNode) *Solution {
	var rangeVals []int
	for head != nil {
		rangeVals = append(rangeVals, head.Val)
		head = head.Next
	}

	return &Solution{rangeVals: rangeVals}
}
