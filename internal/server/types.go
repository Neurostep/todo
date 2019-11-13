package server

import (
	"github.com/Neurostep/todo/pkg/types"
)

type (
	NewTodo struct {
		Title   string        `json:"title" binding:"max=2047"`
		DueDate types.DueDate `json:"due_date"`
	}

	UpdateTodo struct {
		NewTodo
		Done bool `json:"done"`
	}

	NewComment struct {
		Text string `json:"text" binding:"max=2047"`
	}

	NewLabel struct {
		Text  string `json:"text" binding:"max=2047"`
		Color string `json:"color"`
	}

	CommentResponse struct {
		ID   uint   `json:"id"`
		Text string `json:"text"`
	}

	LabelResponse struct {
		ID    uint   `json:"id"`
		Text  string `json:"text"`
		Color string `json:"color"`
	}

	TodoResponse struct {
		ID      uint   `json:"id"`
		Title   string `json:"title"`
		DueDate string `json:"due_date"`
		Done    bool   `json:"done"`
	}

	TodosResponse struct {
		HasMore    bool           `json:"has_more"`
		TotalCount int            `json:"total_count"`
		Data       []TodoResponse `json:"data"`
	}

	TodosQuery struct {
		Limit  uint32 `form:"limit" binding:"lte=1000"`
		Offset uint32 `form:"offset"`
	}
)
