package server

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.opencensus.io/trace"

	"github.com/Neurostep/todo/pkg/services/todo"
	"github.com/Neurostep/todo/pkg/tools/logging"
)

func (r *api) addCommentToTodo(c *gin.Context) {
	ctx, span := trace.StartSpan(c.Request.Context(), "add_comment")
	defer span.End()
	logger := logging.FromContext(ctx, r.logger)

	var req NewComment
	if err := c.ShouldBindJSON(&req); err != nil {
		errs := extractBindErrors(err)
		respondErrors(c, logger, http.StatusBadRequest, errs...)
		return
	}

	idStr := c.Param("id")
	if idStr == "" {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo.comment", "id is empty"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo.comment", "id is not numeric"))
		return
	}

	comment, err := r.conf.TodoService.AddComment(c, r.conf.DB, todo.AddComment{
		TodoId: uint(id),
		Text:   req.Text,
	})

	if err != nil {
		respondErrors(c, logger, http.StatusInternalServerError, newError("todo.comment", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, CommentResponse{
		ID:   comment.ID,
		Text: comment.Text,
	})
}

func (r *api) removeCommentFromTodo(c *gin.Context) {
	ctx, span := trace.StartSpan(c.Request.Context(), "remove_comment")
	defer span.End()
	logger := logging.FromContext(ctx, r.logger)

	idStr := c.Param("id")
	if idStr == "" {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo.comment", "id is empty"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo.comment", "id is not numeric"))
		return
	}

	commentIdStr := c.Param("commentId")
	if commentIdStr == "" {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo.comment", "commentId is empty"))
		return
	}

	commentId, err := strconv.Atoi(commentIdStr)
	if err != nil {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo.comment", "commentId is not numeric"))
		return
	}

	err = r.conf.TodoService.RemoveComment(c, r.conf.DB, uint(id), uint(commentId))

	if err != nil {
		respondErrors(c, logger, http.StatusInternalServerError, newError("todo.comment", err.Error()))
		return
	}

	c.Writer.WriteHeader(http.StatusNoContent)
}

func (r *api) getComments(c *gin.Context) {
	ctx, span := trace.StartSpan(c.Request.Context(), "get_comments")
	defer span.End()
	logger := logging.FromContext(ctx, r.logger)

	idStr := c.Param("id")
	if idStr == "" {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo.comment", "id is empty"))
		return
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		respondErrors(c, logger, http.StatusBadRequest, newError("todo.comment", "id is not numeric"))
		return
	}

	labels, err := r.conf.TodoService.GetComments(c, r.conf.DB, uint(id))

	if err != nil {
		respondErrors(c, logger, http.StatusInternalServerError, newError("todo.comment", err.Error()))
		return
	}

	response := make([]CommentResponse, len(labels))

	for i, label := range labels {
		response[i] = CommentResponse{
			ID:   label.ID,
			Text: label.Text,
		}
	}

	c.JSON(http.StatusOK, response)
}
