package page

import "testing"

type Test struct {
	Id int64
}

func TestXxx(t *testing.T) {
	paging := NewPaging(1)
	p := NewPagingResult[*Test](paging, 10)
	p.Results = &Test{}
}
