package main

import "fmt"

type LinkNode struct {
	next  *LinkNode
	value int
}

func (n *LinkNode) Print() {
	//выведем элементы на экран
	cur := n
	for cur != nil {
		splitter := ""
		if cur != n {
			splitter = "->"
			splitter = "->"
		}
		fmt.Printf("%s%d", splitter, cur.value)
		cur = cur.next
	}

	fmt.Println()
}

// Определим метод обращения списка
func (n *LinkNode) Reverse() *LinkNode {
	//НУжны два указателя на предыдущий и текущий элемент
	var cur = n
	var prev *LinkNode
	for cur != nil {
		//сохраним ссылку на следующий элемент чтобы заменить ее на предыдущий
		next := cur.next
		cur.next = prev
		prev = cur
		cur = next
	}

	return prev
}

func main() {
	//инициализация списка
	n1 := &LinkNode{
		value: 1,
	}
	n2 := &LinkNode{
		value: 2,
	}
	n1.next = n2
	n3 := &LinkNode{
		value: 3,
	}

	n2.next = n3
	//Метод печати в отдельный метод
	n1.Print()
	n1.Reverse().Print()
}
