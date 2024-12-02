package formatTools

import (
	"encoding/json"
	"math/rand"
	"time"
)

// RandGetSlice 切片随机取数
func RandGetSlice[T any](origin []T, count int) []T {
	tmpOrigin := make([]T, len(origin))
	copy(tmpOrigin, origin)
	//一定要seed
	rand.Seed(time.Now().Unix())
	rand.Shuffle(len(tmpOrigin), func(i int, j int) {
		tmpOrigin[i], tmpOrigin[j] = tmpOrigin[j], tmpOrigin[i]
	})

	result := make([]T, 0, count)
	for index, value := range tmpOrigin {
		if index == count {
			break
		}
		result = append(result, value)
	}
	return result
}

// RemoveSliceDuplicate 切片去重复
func RemoveSliceDuplicate[T any](personList []T) (result []T, err error) {
	resultMap := make(map[string]struct{})
	for _, v := range personList {
		data, _ := json.Marshal(v)
		resultMap[string(data)] = struct{}{}
	}

	for k := range resultMap {
		var t T
		err = json.Unmarshal([]byte(k), &t)
		result = append(result, t)
	}

	return
}

// SliceDiff 切片差集
func SliceDiff[T string | int64 | int](a, b []T) []T {
	m := make(map[T]bool)
	for _, v := range b {
		m[v] = true
	}

	result := make([]T, 0)
	for _, v := range a {
		if !m[v] {
			result = append(result, v)
		}
	}

	return result
}
