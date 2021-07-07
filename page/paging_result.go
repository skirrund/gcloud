package page

type PagingResult struct {
	Count     int64       `json:"count"`
	Current   int         `json:"current"`
	PageSize  int         `json:"pageSize"`
	Results   interface{} `json:"results"`
	TotalPage int64       `json:"totalPage"`
}

func (pr *PagingResult) GetTotalPage() int64 {
	if pr.PageSize == 0 {
		pr.PageSize = 10
	}
	return (pr.Count + int64(pr.PageSize) - 1) / int64(pr.PageSize)
}

func (pr *PagingResult) SetTotalPage() {
	pr.TotalPage = pr.GetTotalPage()
}

func NewPagingResult(paging *Paging, count int64) *PagingResult {
	return &PagingResult{
		Current:  paging.Page,
		PageSize: paging.PageSize,
		Count:    count,
	}
}

func NewPagingResult2(paging Paging) *PagingResult {
	return &PagingResult{
		Current:  paging.Page,
		PageSize: paging.PageSize,
	}
}
