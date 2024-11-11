package main

import (
	"checkip/alg"
	"fmt"
)

func main() {
	queryIPv4 := "172.16.254.1"
	queryIPv6 := "2001:0db8:0000:0000:0000:0000:0010:ad12"

	res4 := alg.ValidIPAddress(queryIPv4)
	res6 := alg.ValidIPAddress(queryIPv6)
	fmt.Println(res4)
	fmt.Println(res6)

}
