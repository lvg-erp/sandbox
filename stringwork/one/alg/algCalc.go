package alg

import (
	"strings"
)

// 1
func ReversWords(s string) string {
	words := strings.Fields(s)
	for i := 0; i < len(words)/2; i++ {
		words[i], words[len(words)-i-1] = words[len(words)-i-1], words[i]
	}

	return strings.Join(words, " ")

}

// 2
func backtrackonly(s string,
	wordSet map[string]struct{},
	currentSentence string,
	results *[]string,
	startIndex int) {
	if startIndex == len(s) {
		*results = append(*results, currentSentence)
		return
	}

	for endIndex := startIndex + 1; endIndex <= len(s); endIndex++ {
		word := s[startIndex:endIndex]
		if _, exists := wordSet[word]; exists {
			originalSentence := currentSentence
			if len(currentSentence) > 0 {
				currentSentence += " "
			}
			currentSentence += word
			backtrackonly(s, wordSet, currentSentence, results, endIndex)
			currentSentence = originalSentence
		}
	}

}

func WordBreak(s string, wordDict []string) []string {
	wordSet := make(map[string]struct{})
	for _, word := range wordDict {
		wordSet[word] = struct{}{}
	}
	var results []string
	backtrackonly(s, wordSet, "", &results, 0)

	return results
}

// 3
func backtrackall(s string,
	wordSet map[string]struct{},
	results *[]string,
	startIndex int) {
	if startIndex == len(s) {
		return
	}
	currentSentence := ""
	for endIndex := startIndex + 1; endIndex <= len(s); endIndex++ {
		word := s[startIndex:endIndex]
		if _, exists := wordSet[word]; exists {
			originalSentence := currentSentence
			if len(currentSentence) > 0 {
				currentSentence += " "
			}
			currentSentence += word
			*results = append(*results, currentSentence)

			backtrackall(s, wordSet, results, endIndex)
			currentSentence = originalSentence

		}
	}

}

func WordBreakAll(s string, wordDict []string) []string {
	wordSet := make(map[string]struct{})
	for _, word := range wordDict {
		wordSet[word] = struct{}{}
	}
	var results []string
	backtrackall(s, wordSet, &results, 0)

	return results
}
