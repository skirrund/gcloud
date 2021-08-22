package geohash

import (
	"fmt"
	"testing"
)

//返回单位为：米
func Test_Distance(testing *testing.T) {
	// 121.464254, 31.197407        121.490904,31.17268
	fmt.Println(Distance(31.197407, 31.17268, 121.464254, 121.490904))
}
