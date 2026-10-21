package cache

import "sort"

// FilterSortedPage 在一次读快照中统计全部匹配项，只保留目标页之前的有界候选。
// less 必须提供稳定全序，且不得修改/保留参数指针；predicate/less 不得回调当前缓存。浅分页使用最大堆，
// 深分页退化为普通排序，避免堆维护反而增加开销；返回页不持有全量结果数组。
func (c *MapCache[K, V]) FilterSortedPage(predicate func(V) bool, less func(*V, *V) bool, offset, limit int) ([]V, int) {
	if offset < 0 {
		offset = 0
	}
	var items []V
	total := 0
	func() {
		c.lock.RLock()
		defer c.lock.RUnlock()
		size := len(c.data)
		keep := 0
		if limit > 0 && offset < size {
			keep = offset + min(limit, size-offset) // 避免 offset+limit 溢出。
		}
		useHeap := keep > 0 && keep < size/2
		capacity := keep
		if keep > 0 && !useHeap {
			capacity = size
		}
		items = make([]V, 0, capacity)
		var v V // 复用候选槽，比较器取地址时不为每个节点分配对象。
		for _, v = range c.data {
			if !predicate(v) {
				continue
			}
			total++
			if keep == 0 {
				continue
			}
			if !useHeap {
				items = append(items, v)
				continue
			}
			if len(items) < keep {
				items = append(items, v)
				if len(items) == keep {
					for i := len(items)/2 - 1; i >= 0; i-- {
						siftPageDown(items, i, less)
					}
				}
			} else if less(&v, &items[0]) {
				items[0] = v
				siftPageDown(items, 0, less)
			}
		}
	}()
	if offset >= len(items) || limit <= 0 {
		return []V{}, total
	}
	sort.Slice(items, func(i, j int) bool { return less(&items[i], &items[j]) })
	end := offset + min(limit, len(items)-offset)
	result := make([]V, end-offset)
	copy(result, items[offset:end])
	return result, total
}

// 堆顶始终是当前保留项中排序最靠后的值。
func siftPageDown[V any](items []V, root int, less func(*V, *V) bool) {
	for {
		child := root*2 + 1
		if child >= len(items) {
			return
		}
		if child+1 < len(items) && less(&items[child], &items[child+1]) {
			child++
		}
		if !less(&items[root], &items[child]) {
			return
		}
		items[root], items[child] = items[child], items[root]
		root = child
	}
}

// IndexValueCounts 返回各索引值的节点数快照，不复制实体内容。
func (c *MapCache[K, V]) IndexValueCounts(field string) map[string]int {
	c.lock.RLock()
	defer c.lock.RUnlock()
	result := make(map[string]int)
	if index, ok := c.indexes[field]; ok {
		for value, keys := range index.data {
			result[value] = len(keys)
		}
	}
	return result
}
