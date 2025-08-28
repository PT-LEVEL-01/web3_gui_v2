package utils

/*
	去除重复，不保证顺序
*/
func DistinctString(list []string) []string {
	results := make([]string, 0)
	m := make(map[string]int)
	for _, one := range list {
		_, ok := m[one]
		if ok {
			continue
		}
		m[one] = 0
		results = append(results, one)
	}
	return results
}
