package pagination

import (
	"math"
	"strconv"
)

// PaginationResponse represents the pagination response structure similar to Laravel
type PaginationResponse[T any] struct {	
	CurrentPage  int         `json:"current_page"`
	Data         []T `json:"data"`
	FirstPageURL string      `json:"first_page_url"`
	From         int         `json:"from"`
	LastPage     int         `json:"last_page"`
	LastPageURL  string      `json:"last_page_url"`
	NextPageURL  string      `json:"next_page_url"`
	Path         string      `json:"path"`
	PerPage      int         `json:"per_page"`
	PrevPageURL  string      `json:"prev_page_url"`
	To           int         `json:"to"`
	Total        int         `json:"total"`
}

// Paginate creates a pagination response from the given data
func Paginate[T any](data []T, total int, page int, perPage int, path string) PaginationResponse[T] {
	lastPage := int(math.Ceil(float64(total) / float64(perPage)))
	
	from := (page-1)*perPage + 1
	to := from + perPage - 1
	if to > total {
		to = total
	}
	
	var prevPageURL, nextPageURL string
	if page > 1 {
		prevPageURL = path + "?page=" + strconv.Itoa(page-1)
	}
	if page < lastPage {
		nextPageURL = path + "?page=" + strconv.Itoa(page+1)
	}

	return PaginationResponse[T]{
		CurrentPage:  page,
		Data:         data,
		FirstPageURL: path + "?page=1",
		From:         from,
		LastPage:     lastPage,
		LastPageURL:  path + "?page=" + strconv.Itoa(lastPage),
		NextPageURL:  nextPageURL,
		Path:         path,
		PerPage:      perPage,
		PrevPageURL:  prevPageURL,
		To:           to,
		Total:        total,
	}
}
