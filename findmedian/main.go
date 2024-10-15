package main

import (
	"findmedian/alg"
	"fmt"
)

func main() {

	medianFinder := alg.NewMedianFinder()
	medianFinder.AddNum(1)
	medianFinder.AddNum(2)
	medianFinder.AddNum(4)
	medianFinder.AddNum(6)

	res := medianFinder.FindMedian()
	fmt.Println(res)

}
