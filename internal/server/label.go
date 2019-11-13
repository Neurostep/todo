package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.opencensus.io/trace"

	"github.com/Neurostep/todo/pkg/services/todo"
	"github.com/Neurostep/todo/pkg/tools/logging"
)

func (r *api) addLabelToTodo(c *gin.Context) {
	ctx, span := trace.StartSpan(c.Request.Context(), "add_label")
	defer span.End()
	logger := logging.FromContext(ctx, r.logger)

	var req NewLabel
	if err := c.ShouldBindJSON(&req); err != nil {
		errs := extractBindErrors(err)
		respondErrors(c, logger, http.StatusBadRequest, errs...)
		return
	}

	idStr := c.Param("id")
	if idStr == "" {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo.label", "id is empty"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo.label", "id is not numeric"))
		return
	}

	label, err := r.conf.TodoService.AddLabel(c, r.conf.DB, todo.AddLabel{
		TodoId: uint(id),
		Color:  req.Color,
		Text:   req.Text,
	})

	if err != nil {
		respondErrors(c, logger, http.StatusInternalServerError, newError("todo.label", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, LabelResponse{
		ID:    label.ID,
		Text:  label.Text,
		Color: label.Color,
	})
}

func (r *api) removeLabelFromTodo(c *gin.Context) {
	ctx, span := trace.StartSpan(c.Request.Context(), "remove_label")
	defer span.End()
	logger := logging.FromContext(ctx, r.logger)

	idStr := c.Param("id")
	if idStr == "" {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo.label", "id is empty"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo.label", "id is not numeric"))
		return
	}

	labelIdStr := c.Param("labelId")
	if labelIdStr == "" {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo.label", "labelId is empty"))
		return
	}

	labelId, err := strconv.Atoi(labelIdStr)
	if err != nil {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo.label", "labelId is not numeric"))
		return
	}

	err = r.conf.TodoService.RemoveLabel(c, r.conf.DB, uint(id), uint(labelId))

	if err != nil {
		respondErrors(c, logger, http.StatusInternalServerError, newError("todo.label", err.Error()))
		return
	}

	c.Writer.WriteHeader(http.StatusNoContent)
}

func (r *api) getLabels(c *gin.Context) {
	ctx, span := trace.StartSpan(c.Request.Context(), "get_labels")
	defer span.End()
	logger := logging.FromContext(ctx, r.logger)

	idStr := c.Param("id")
	if idStr == "" {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo.label", "id is empty"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo.label", "id is not numeric"))
		return
	}

	labels, err := r.conf.TodoService.GetLabels(c, r.conf.DB, uint(id))

	if err != nil {
		respondErrors(c, logger, http.StatusInternalServerError, newError("todo.label", err.Error()))
		return
	}

	response := make([]LabelResponse, len(labels))

	for i, label := range labels {
		response[i] = LabelResponse{
			ID:    label.ID,
			Text:  label.Text,
			Color: label.Color,
		}
	}

	c.JSON(http.StatusOK, response)
}
