package utils

import (
	"fmt"
	"testing"
)

func Test_Append(testing *testing.T)  {

	slice := []string{"11","xx","yy"}
	fmt.Println(slice)
	slice = AppendStr(slice, "22", true)
	fmt.Println(slice)
	slice = AppendStr(slice, "xx", true)
	fmt.Println(slice)
}
