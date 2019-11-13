package server

import (
	"github.com/Neurostep/todo/pkg/types"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.opencensus.io/trace"

	"github.com/Neurostep/todo/pkg/services/todo"
	"github.com/Neurostep/todo/pkg/tools/logging"
)

func (r *api) getTodos(c *gin.Context) {
	ctx, span := trace.StartSpan(c.Request.Context(), "list_todos")
	defer span.End()
	logger := logging.FromContext(ctx, r.logger)

	query := TodosQuery{}
	if err := c.ShouldBindQuery(&query); err != nil {
		errs := extractBindErrors(err)
		respondErrors(c, logger, http.StatusBadRequest, errs...)
		return
	}

	results, err := r.conf.TodoService.GetTodos(ctx, r.conf.DB, todo.PaginateTodos{
		Limit:  query.Limit,
		Offset: query.Offset,
	})

	if err != nil {
		respondErrors(c, logger, http.StatusInternalServerError)
		return
	}

	res := TodosResponse{
		HasMore:    results.HasMore,
		TotalCount: results.TotalCount,
		Data:       make([]TodoResponse, 0, len(results.Items)),
	}

	for _, e := range results.Items {
		res.Data = append(res.Data, TodoResponse{
			ID:      e.ID,
			Title:   e.Title,
			Done:    e.Done,
			DueDate: e.DueDate.Format(types.DueDateFormat),
		})
	}
	c.JSON(http.StatusOK, res)
}

func (r *api) getTodo(c *gin.Context) {
	ctx, span := trace.StartSpan(c.Request.Context(), "get_todo")
	defer span.End()
	logger := logging.FromContext(ctx, r.logger)

	idStr := c.Param("id")
	if idStr == "" {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo", "id is empty"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo", "id is not numeric"))
		return
	}

	td, err := r.conf.TodoService.GetTodo(ctx, r.conf.DB, uint(id))
	if err != nil {
		respondErrors(c, logger, http.StatusInternalServerError, newError("todo", err.Error()))
		return
	}

	c.JSON(http.StatusOK, TodoResponse{
		ID:      td.ID,
		Title:   td.Title,
		Done:    td.Done,
		DueDate: td.DueDate.Format(types.DueDateFormat),
	})
}

func (r *api) createTodo(c *gin.Context) {
	ctx, span := trace.StartSpan(c.Request.Context(), "create_todo")
	defer span.End()
	logger := logging.FromContext(ctx, r.logger)

	var req NewTodo
	if err := c.ShouldBindJSON(&req); err != nil {
		errs := extractBindErrors(err)
		respondErrors(c, logger, http.StatusBadRequest, errs...)
		return
	}

	td, err := r.conf.TodoService.CreateTodo(ctx, r.conf.DB, &todo.CreateTodo{
		Title:   req.Title,
		DueDate: req.DueDate,
	})
	if err != nil {
		respondErrors(c, logger, http.StatusInternalServerError, newError("todo", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, TodoResponse{
		ID:      td.ID,
		Title:   td.Title,
		Done:    td.Done,
		DueDate: td.DueDate.Format(types.DueDateFormat),
	})
}

func (r *api) updateTodo(c *gin.Context) {
	ctx, span := trace.StartSpan(c.Request.Context(), "update_todo")
	defer span.End()
	logger := logging.FromContext(ctx, r.logger)

	idStr := c.Param("id")
	if idStr == "" {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo", "id is empty"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo", "id is not numeric"))
		return
	}

	var req UpdateTodo
	if err := c.ShouldBindJSON(&req); err != nil {
		errs := extractBindErrors(err)
		respondErrors(c, logger, http.StatusBadRequest, errs...)
		return
	}

	td, err := r.conf.TodoService.UpdateTodo(c, r.conf.DB, &todo.UpdateTodo{
		Id:      uint(id),
		Title:   req.Title,
		DueDate: req.DueDate,
		Done:    req.Done,
	})

	if err != nil {
		respondErrors(c, logger, http.StatusInternalServerError, newError("todo", err.Error()))
		return
	}

	c.JSON(http.StatusOK, TodoResponse{
		ID:      td.ID,
		Title:   td.Title,
		Done:    td.Done,
		DueDate: td.DueDate.Format(types.DueDateFormat),
	})
}

func (r *api) deleteTodo(c *gin.Context) {
	ctx, span := trace.StartSpan(c.Request.Context(), "delete_todo")
	defer span.End()
	logger := logging.FromContext(ctx, r.logger)

	idStr := c.Param("id")
	if idStr == "" {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo", "id is empty"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo", "id is not numeric"))
		return
	}

	err = r.conf.TodoService.DeleteTodo(c, r.conf.DB, uint(id))

	if err != nil {
		respondErrors(c, logger, http.StatusInternalServerError, newError("todo", err.Error()))
		return
	}

	c.Writer.WriteHeader(http.StatusNoContent)
}
