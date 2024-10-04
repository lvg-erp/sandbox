package alg

import (
	"container/list"
	"sort"
)

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func InsertLevelOrder(arr []int, root *TreeNode, i int) *TreeNode {
	// Если индекс выходит за пределы массива, возвращаем nil
	if i < len(arr) {
		if arr[i] == 0 {
			return nil
		}
		// Создаем новый узел
		root = &TreeNode{Val: arr[i]}
		// Рекурсивно создаем левого и правого ребенка
		root.Left = InsertLevelOrder(arr, root.Left, 2*i+1)   // Для левого поддерева
		root.Right = InsertLevelOrder(arr, root.Right, 2*i+2) // Для правого поддерева
	}
	return root
}

func VerticalOrder(root *TreeNode) [][]int {
	var output [][]int
	if root == nil {
		return output
	}

	columnTable := map[int][]int{}
	quene := list.New()
	quene.PushBack([2]interface{}{root, 0})
	for quene.Len() > 0 {
		element := quene.Front()
		quene.Remove(element)
		pair := element.Value.([2]interface{})
		node := pair[0].(*TreeNode)
		column := pair[1].(int)

		if node != nil {
			columnTable[column] = append(columnTable[column], node.Val)

			if node.Left != nil {
				quene.PushBack([2]interface{}{node.Left, column - 1})
			}
			if node.Right != nil {
				quene.PushBack([2]interface{}{node.Right, column + 1})
			}

		}
	}

	var sortedKeys []int
	for key := range columnTable {
		sortedKeys = append(sortedKeys, key)
	}
	sort.Ints(sortedKeys)

	for _, key := range sortedKeys {
		output = append(output, columnTable[key])
	}

	return output
}
