package alg

import (
	treenode "github.com/egregors/TreeNode"
)

func InvertingTree(root *treenode.TreeNode) *treenode.TreeNode {
	if root == nil {
		return nil
	}

	right := InvertingTree(root.Right)
	left := InvertingTree(root.Left)
	root.Left = right
	root.Right = left
	return root
}
