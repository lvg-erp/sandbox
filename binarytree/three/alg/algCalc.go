package alg

type TreeNode struct {
	Val   int
	Left  *TreeNode
	Right *TreeNode
}

func CreateTreeFormArray(arr []int, root *TreeNode, i int) *TreeNode {

	if i < len(arr) {
		if arr[i] == 0 {
			return nil
		}

		//Новый узел
		root = &TreeNode{Val: arr[i]}
		root.Left = CreateTreeFormArray(arr, root.Left, 2*i+1)
		root.Right = CreateTreeFormArray(arr, root.Right, 2*i+2)
	}

	return root
}

func MaxPathSum(root *TreeNode) (int, []int) {
	if root == nil {
		return 0, nil
	}

	leftSum, leftPath := MaxPathSum(root.Left)
	rightSum, rightPath := MaxPathSum(root.Right)

	if leftSum > rightSum {
		return root.Val + leftSum, append([]int{root.Val}, leftPath...)
	} else {
		return root.Val + rightSum, append([]int{root.Val}, rightPath...)
	}

}
