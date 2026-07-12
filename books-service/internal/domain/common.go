package domain

// Pagination метаданные пагинации для списков
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ErrorResponse единый формат ошибки API
type ErrorResponse struct {
	Code      string        `json:"code"`
	Message   string        `json:"message"`
	Details   []ErrorDetail `json:"details,omitempty"`
	RequestID string        `json:"request_id"`
}

// ErrorDetail детальная информация об ошибке валидации
type ErrorDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ListParams struct {
	Search string
	Sort   string
	Order  string
	Page   int
	Limit  int
}

// NewPagination создаёт структуру пагинации на основе параметров запроса
func NewPagination(page, limit, total int) Pagination {
	p := Pagination{
		Page:  page,
		Limit: limit,
		Total: total,
	}

	if limit > 0 {
		p.TotalPages = (total + limit - 1) / limit
	}

	return p
}