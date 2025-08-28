package response

import (
	"digital-book-lending/utils"
	"math"
	"net/http"

	"github.com/google/uuid"
)

// Success is an alias for Api for swag documentation.
type Success Api

// Error is an alias for Api for swag documentation.
type Error Api

// Pagination is an alias for PaginatedResponse for swag documentation.
type Pagination PaginatedResponse

type Errors struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

type Api struct {
	Id      uuid.UUID   `json:"log_id"`
	Code    int         `json:"code,omitempty"`
	Status  bool        `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Errors  interface{} `json:"errors,omitempty"`
}

type PaginatedResponse struct {
	LogID       string      `json:"log_id"`
	Code        int         `json:"code"`
	Status      string      `json:"status"`
	Message     string      `json:"message"`
	TotalData   int         `json:"total_data"`
	TotalPages  int         `json:"total_pages"`
	CurrentPage int         `json:"current_page"`
	NextPage    bool        `json:"next_page"`
	PrevPage    bool        `json:"prev_page"`
	Limit       int         `json:"limit"`
	Data        interface{} `json:"data"`
	Error       interface{} `json:"error,omitempty"`
}

func Response(code int, msg string, logId uuid.UUID, data interface{}) *Api {
	res := new(Api)
	res.Id = logId
	res.Message = msg
	res.Data = data

	switch code {
	case http.StatusOK, http.StatusCreated:
		res.Status = true
	default:
		res.Status = false
	}

	return res
}

func PaginationResponse(code, total, page, perPage int, logId uuid.UUID, data interface{}) *PaginatedResponse {
	res := new(PaginatedResponse)

	// Count total pages
	var totalPages int
	if total > 0 && perPage > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(perPage)))
	} else if total > 0 {
		totalPages = 1
	}

	// Check for next page (hasNext)
	hasNext := false
	if page < totalPages {
		hasNext = true
	}

	message := utils.MsgSuccess
	if total == 0 || page > totalPages {
		message = utils.MsgNotFound
	}

	res.LogID = logId.String()
	res.Code = code
	res.Status = http.StatusText(code)
	res.Message = message
	res.Data = data
	res.TotalData = total
	res.TotalPages = totalPages
	res.CurrentPage = page
	res.NextPage = hasNext
	res.PrevPage = page > 1
	res.Limit = perPage

	return res
}
