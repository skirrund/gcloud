package utils

import (
	"fmt"
	"testing"
)

func Test_Append(testing *testing.T) {
	str := "12.3中r.pdf"
	idx := UnicodeLastIndex(str, ".")
	fmt.Println(idx)
}
