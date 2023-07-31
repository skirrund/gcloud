package page

type PagingResult[T any] struct {
	Count     int64 `json:"count"`
	Current   int   `json:"current"`
	PageSize  int   `json:"pageSize"`
	Results   T     `json:"results"`
	TotalPage int64 `json:"totalPage"`
}

func (pr *PagingResult[T]) GetTotalPage() int64 {
	if pr.PageSize == 0 {
		pr.PageSize = 10
	}
	return (pr.Count + int64(pr.PageSize) - 1) / int64(pr.PageSize)
}

func (pr *PagingResult[T]) SetTotalPage() {
	pr.TotalPage = pr.GetTotalPage()
}

func NewPagingResult[T any](paging Paging, count int64) *PagingResult[T] {
	return &PagingResult[T]{
		Current:  paging.Page,
		PageSize: paging.PageSize,
		Count:    count,
	}
}

func NewPagingResult2[T any](paging Paging) *PagingResult[T] {
	return &PagingResult[T]{
		Current:  paging.Page,
		PageSize: paging.PageSize,
	}
}
