package utils

import "math"

// ComputePageCount computes pages needed to paginate in order to get the count of items.
func ComputePageCount(count int, countPerPage int) int {
	pages := int(math.Ceil(float64(count) / float64(countPerPage)))
	if pages < 1 {
		pages = 1
	}
	return pages
}
