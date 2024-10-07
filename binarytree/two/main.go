package main

import (
	"binarytreetwo/alg"
	"fmt"
	tn "github.com/egregors/TreeNode"
	"log"
)

func main() {
	data := "[4,2,7,1,3,6,9]"
	root, err := tn.NewTreeNode(data)
	if err != nil {
		log.Fatal(err)
	}

	res := alg.InvertingTree(root)
	fmt.Println(res)

}
