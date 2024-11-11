package alg

import (
	"strconv"
	"strings"
	"unicode"
)

func validateIP4(ip string) bool {
	nums := strings.Split(ip, ".")
	if len(nums) != 4 {
		return false
	}
	for _, n := range nums {
		if len(n) == 0 || len(n) > 3 {
			return false
		}
		for _, ch := range n {
			if !unicode.IsDigit(ch) {
				return false
			}
		}
		num, err := strconv.Atoi(n)
		if err != nil || num > 255 {
			return false
		}
	}

	return true
}

func validateIP6(ip string) bool {
	nums := strings.Split(ip, ":")
	hexdigits := "0123456789abcdefABCDEF"
	if len(nums) != 8 {
		return false
	}
	for _, n := range nums {
		if len(n) == 0 || len(n) > 4 {
			return false
		}
		for _, ch := range n {
			if !strings.ContainsRune(hexdigits, ch) {
				return false
			}
		}
	}

	return true
}

func ValidIPAddress(ip string) string {
	if strings.Count(ip, ".") == 3 {
		if validateIP4(ip) {
			return "IPv4 correction"
		} else {
			return "IPv4 not correction"
		}
	} else if strings.Count(ip, ":") == 7 {
		if validateIP6(ip) {
			return "IPv6 correction"
		} else {
			return "IPv6 not correction"
		}
	} else {
		return "Not correction format IP adressess"
	}
}
