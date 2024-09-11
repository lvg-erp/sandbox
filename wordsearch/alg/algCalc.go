package alg

type TrieNode struct {
	children map[rune]*TrieNode
	word     string
}

type Solution struct {
	board  [][]byte
	result []string
}

func FindWords(board [][]byte, words []string) []string {
	root := buildTrie(words)
	sol := Solution{board: board}

	for row := range board {
		for col := range board[0] {
			if _, ok := root.children[rune(board[row][col])]; ok {
				sol.backtrack(row, col, root)
			}
		}
	}
	return sol.result
}

func buildTrie(words []string) *TrieNode {
	root := &TrieNode{children: make(map[rune]*TrieNode)}
	for _, word := range words {
		node := root
		for _, letter := range word {
			if node.children[letter] == nil {
				node.children[letter] = &TrieNode{children: make(map[rune]*TrieNode)}
			}
			node = node.children[letter]
		}
		node.word = word
	}
	return root
}

func (sol *Solution) backtrack(row, col int, parent *TrieNode) {
	letter := rune(sol.board[row][col])
	currentNode := parent.children[letter]

	if currentNode.word != "" {
		sol.result = append(sol.result, currentNode.word)
		currentNode.word = ""
	}

	sol.board[row][col] = '#'
	rowOffset := []int{-1, 0, 1, 0}
	colOffset := []int{0, 1, 0, -1}

	for i := 0; i < 4; i++ {
		newRow := row + rowOffset[i]
		newCol := col + colOffset[i]
		if newRow >= 0 && newRow < len(sol.board) && newCol >= 0 && newCol < len(sol.board[0]) {
			if _, ok := currentNode.children[rune(sol.board[newRow][newCol])]; ok {
				sol.backtrack(newRow, newCol, currentNode)
			}
		}
	}

	sol.board[row][col] = byte(letter)
	if len(currentNode.children) == 0 {
		delete(parent.children, letter)
	}

}
