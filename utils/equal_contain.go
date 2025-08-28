package utils

// 对比两个数组是否相等
func EqualSliceFunc[T any](find *[]T, slice *[]T, f func(a, b *T) bool) bool {
	if find == nil || slice == nil {
		return false
	}
	ok := false
	for _, i := range *find {
		for _, j := range *slice {
			ok = f(&i, &j)
			if !ok {
				return false
			}
		}
	}
	return true
}

// 对比两个数组是否相等
func EqualMapFunc[T any](find *map[string]T, slice *map[string]T, f func(a, b *T) bool) bool {
	if find == nil || slice == nil {
		return false
	}
	if len(*find) != len(*slice) {
		return false
	}
	ok := false
	var v2 T
	for k, v := range *find {
		v2, ok = (*slice)[k]
		if !ok {
			return false
		}
		ok = f(&v, &v2)
		if !ok {
			return false
		}
	}
	return true
}

// 在数组中查找是否包含元素
func ContainSliceFunc[T any](slice *[]T, f func(one *T) bool) (bool, *T) {
	if slice == nil {
		return false, nil
	}
	have := false
	for _, one := range *slice {
		have = f(&one)
		if have {
			return true, &one
		}
	}
	return false, nil
}
