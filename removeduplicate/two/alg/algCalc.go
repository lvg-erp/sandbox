package alg

import "fmt"

type ListNode struct {
	Val  int
	Next *ListNode
}

type LinkedList struct {
	head *ListNode
}

//func NewLinkedList(head *ListNode) *LinkedList {
//	return &LinkedList{head: head}
//}

func DeleteDuplicates(head *ListNode) *ListNode {
	sentinel := &ListNode{
		0,
		head,
	}
	pred := sentinel
	for head != nil {
		if head.Next != nil && head.Val == head.Next.Val {
			for head.Next != nil && head.Val == head.Next.Val {
				head = head.Next
			}
			pred.Next = head.Next
		} else {
			pred = pred.Next
		}

		head = head.Next
	}

	return sentinel.Next
}

func PrintLinkedList(head *ListNode) {
	curr := head

	for curr != nil {
		fmt.Printf("%d", curr.Val)
		curr = curr.Next
	}

	fmt.Println()
}

// Вставка в список после первого узла
func (list *LinkedList) InsertAtFront(data int) {
	if list.head == nil {
		newNode := &ListNode{Val: data, Next: nil}
		list.head = newNode
		return
	}

	newNode := &ListNode{
		Val:  data,
		Next: nil,
	}
	list.head = newNode
}

// Вставка в список после последнего узла
func (list *LinkedList) InsertAtBack(data int) {
	newNode := &ListNode{Val: data, Next: nil}

	if list.head == nil {
		list.head = newNode
		return
	}

	current := list.head
	for current.Next != nil {
		current = current.Next
	}
	current.Next = newNode

}
