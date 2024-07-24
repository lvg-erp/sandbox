package main

import "removed/alg"

func main() {

	node1 := &alg.ListNode{Val: 1}
	node2 := &alg.ListNode{Val: 2}
	node3 := &alg.ListNode{Val: 3}
	node4 := &alg.ListNode{Val: 3}
	node5 := &alg.ListNode{Val: 4}
	node6 := &alg.ListNode{Val: 4}
	node7 := &alg.ListNode{Val: 5}
	node1.Next = node2
	node2.Next = node3
	node3.Next = node4
	node4.Next = node5
	node5.Next = node6
	node6.Next = node7

	//ln:=alg.ListNode{}
	//alg.NewLinkedList(&ln)
	//ln:=alg.NewListNode(0, nil)
	//ll := alg.LinkedList{}
	//ll.InsertAtFront(10)
	//ll.InsertAtFront(20)
	//ll.InsertAtFront(30)
	//ll.InsertAtFront(40)
	res := alg.DeleteDuplicates(node1)
	alg.PrintLinkedList(res)

}
