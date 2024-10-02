package alg

func NumWays(n, k int) int {
	if n == 1 {
		return k
	}

	twoPostsBack := k
	onePostBack := k * k
	//3 - колич столбцов окрашенных подряд
	//не может быть больше
	for i := 3; i <= n; i++ {
		curr := (k - 1) * (onePostBack + twoPostsBack)
		twoPostsBack = onePostBack
		onePostBack = curr
	}

	return onePostBack
}
