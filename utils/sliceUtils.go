package utils


func RemoveDuplicateStr(list []string) []string {
	result := make([]string, 0, len(list))
	temp := map[string]struct{}{}
	for _, item := range list {
		if _, ok := temp[item]; !ok { //如果字典中找不到元素，ok=false，!ok为true，就往切片中append元素。
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
func RemoveDuplicateInt64(list []int64) []int64 {
	result := make([]int64, 0, len(list))
	temp := map[int64]struct{}{}
	for _, item := range list {
		if _, ok := temp[item]; !ok { //如果字典中找不到元素，ok=false，!ok为true，就往切片中append元素。
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}
func RemoveDuplicateInt(list []int) []int {
	result := make([]int, 0, len(list))
	temp := map[int]struct{}{}
	for _, item := range list {
		if _, ok := temp[item]; !ok { //如果字典中找不到元素，ok=false，!ok为true，就往切片中append元素。
			temp[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

func AppendStr(list []string, target string, ifAbsent bool) []string {
	if !ifAbsent {
		return append(list, target)
	}
	exists := false
	for _, obj := range list {
		if obj == target {
			exists = true
			break
		}
	}
	if !exists {
		return append(list, target)
	}
	return list
}

func AppendInt64(list []int64, target int64, ifAbsent bool) []int64 {
	if !ifAbsent {
		return append(list, target)
	}
	exists := false
	for _, obj := range list {
		if obj == target {
			exists = true
			break
		}
	}
	if !exists {
		return append(list, target)
	}
	return list
}

func AppendInt(list []int, target int, ifAbsent bool) []int {
	if !ifAbsent {
		return append(list, target)
	}
	exists := false
	for _, obj := range list {
		if obj == target {
			exists = true
			break
		}
	}
	if !exists {
		return append(list, target)
	}
	return list
}