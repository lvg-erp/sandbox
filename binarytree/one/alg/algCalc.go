package alg

import (
	"fmt"
	"github.com/egregors/TreeNode"
)

func constructorPath(root *treenode.TreeNode, path string, paths *[]string) {
	if root != nil {
		path += fmt.Sprintf("%d", root.Val)
		if root.Left == nil && root.Right == nil {
			*paths = append(*paths, path)
		} else {
			path += "->"
			constructorPath(root.Left, path, paths)
			constructorPath(root.Right, path, paths)
		}
	}
}

func BinaryTreePaths(root *treenode.TreeNode) []string {
	var paths []string
	constructorPath(root, "", &paths)
	return paths
}
