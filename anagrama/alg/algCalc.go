package alg

import "strings"

func IsAnagram(first, second string) bool {

	if len(first) != len(second) {
		return false
	}

	perv := strings.Split(first, "")
	vtor := strings.Split(second, "")
	z := 0

	for i := 0; i < len(perv); i++ {
		z = 0
		for j := 0; j < len(vtor) && z == 0; j++ {
			if perv[i] == vtor[j] {
				vtor[j] = ""
				z = 1
			}
		}
		if z == 0 {
			return false
		}
	}
	for l := 0; l < len(vtor); l++ {
		if vtor[l] != "" {
			return false
		}
	}

	return true

}

// для английского алфавита где 26 букв
func IsAnagram_(s, t string) bool {
	if len(s) != len(t) {
		return false
	}

	arr := [26]int{}
	for i := 0; i < len(s); i++ {
		arr[s[i]-'a']++
		arr[t[i]-'a']--
	}
	for i := 0; i < 26; i++ {
		if arr[i] != 0 {
			return false
		}
	}
	return true
}
