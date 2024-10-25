package alg

type LFUCache struct {
	cache       map[int][2]int
	frequencies map[int][]int
	capacity    int
	minf        int
}

func NewLFUCache(capacity int) *LFUCache {
	return &LFUCache{
		cache:       make(map[int][2]int),
		frequencies: make(map[int][]int),
		capacity:    capacity,
	}
}

func (c *LFUCache) insert(key, freq, value int) {
	c.frequencies[freq] = append(c.frequencies[freq], key)
	c.cache[key] = [2]int{freq, value}
}

func (c *LFUCache) Get(key int) int {
	if val, found := c.cache[key]; !found {
		return -1
	} else {
		freq, value := val[0], val[1]
		c.frequencies[freq] = removeKey(c.frequencies[freq], key)
		if len(c.frequencies[freq]) == 0 {
			delete(c.frequencies, freq)
			if c.minf == freq {
				c.minf++
			}
		}

		c.insert(key, freq+1, value)
		return value
	}
}

func (c *LFUCache) Put(key, value int) {
	if c.capacity <= 0 {
		return
	}

	if val, found := c.cache[key]; found {
		c.cache[key] = [2]int{val[0], value}
		c.Get(key)
		return
	}
	if len(c.cache) == c.capacity {
		minFreqKeys := c.frequencies[c.minf]
		delete(c.cache, minFreqKeys[0])
		c.frequencies[c.minf] = minFreqKeys[1:]
		if len(c.frequencies[c.minf]) == 0 {
			delete(c.frequencies, c.minf)
		}
	}
	c.minf = 1
	c.insert(key, 1, value)
}

func removeKey(slice []int, key int) []int {
	for i, v := range slice {
		if v == key {
			return append(slice[:i], slice[i+1:]...)
		}
	}

	return slice
}
