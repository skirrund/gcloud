package page

type Paging struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
}

const (
	PAGE      = 1
	PAGE_SIZE = 10
)

func NewPaging(page int) Paging {
	return NewPaging2(page, PAGE_SIZE)
}

func NewPaging2(page int, pageSize int) Paging {
	p := PAGE
	if page > 0 {
		p = page
	}
	ps := PAGE_SIZE
	if pageSize > 0 {
		ps = pageSize
	}
	return Paging{
		Page:     p,
		PageSize: ps,
	}
}

func (p Paging) GetOffset() int {
	return (p.Page - 1) * p.PageSize
}
