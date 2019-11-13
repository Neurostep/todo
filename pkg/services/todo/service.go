package todo

import (
	"context"
	"time"

	"github.com/Neurostep/todo/pkg/database"
	"github.com/Neurostep/todo/pkg/tools/logging"
	"github.com/Neurostep/todo/pkg/types"
	"github.com/go-kit/kit/log"
	"github.com/jinzhu/gorm"
	"go.opencensus.io/trace"
)

const (
	DefaultMaxLimit = 1000
	DefaultLimit    = 20
	MaxComments     = 5
	MaxLabels       = 10
)

type (
	Config struct {
		DB     *gorm.DB
		Logger log.Logger
	}

	CreateTodo struct {
		Title   string
		DueDate types.DueDate
	}

	UpdateTodo struct {
		Id      uint
		Title   string
		DueDate types.DueDate
		Done    bool
	}

	PaginateTodos struct {
		Offset, Limit uint32
	}

	PaginatedTodos struct {
		HasMore    bool
		TotalCount int
		Items      []Todo
	}

	AddComment struct {
		TodoId uint
		Text   string
	}

	AddLabel struct {
		TodoId uint
		Color  string
		Text   string
	}

	ServiceProvider interface {
		CreateTodo(ctx context.Context, db *gorm.DB, todo *CreateTodo) (*Todo, error)
		UpdateTodo(ctx context.Context, db *gorm.DB, todo *UpdateTodo) (*Todo, error)
		GetTodo(ctx context.Context, db *gorm.DB, id int) (*Todo, error)
		DeleteTodo(ctx context.Context, db *gorm.DB, id int) error
		GetTodos(ctx context.Context, db *gorm.DB, pg PaginateTodos) (*PaginatedTodos, error)
		AddComment(ctx context.Context, db *gorm.DB, comment AddComment) (*Comment, error)
		RemoveComment(ctx context.Context, db *gorm.DB, todoId, id int) error
		AddLabel(ctx context.Context, db *gorm.DB, comment AddLabel) (*Label, error)
		RemoveLabel(ctx context.Context, db *gorm.DB, todoId, id int) error
	}

	Service struct {
		DB     *gorm.DB
		Logger log.Logger
	}
)

func New(cfg Config) *Service {
	return &Service{
		Logger: cfg.Logger,
		DB:     cfg.DB,
	}
}

func (s *Service) CreateTodo(ctx context.Context, db *gorm.DB, todo *CreateTodo) (*Todo, error) {
	ctx, span := trace.StartSpan(ctx, "todo.create")
	defer span.End()
	logger := logging.FromContext(ctx, s.Logger)

	td := &Todo{
		Title:   todo.Title,
		DueDate: *todo.DueDate.Time(),
		Done:    false,
	}

	err := db.Save(td).Error

	if err != nil {
		logger.Log("event", "failed to create todo", "error", err)
		return nil, err
	}

	return td, nil
}

func (s *Service) UpdateTodo(ctx context.Context, db *gorm.DB, todo *UpdateTodo) (*Todo, error) {
	ctx, span := trace.StartSpan(ctx, "todo.update")
	defer span.End()
	logger := logging.FromContext(ctx, s.Logger)

	td := &Todo{
		ID:      todo.Id,
		Title:   todo.Title,
		DueDate: *todo.DueDate.Time(),
		Done:    todo.Done,
	}

	err := db.Scopes(withTodoID(todo.Id)).Save(td).Error

	if err != nil {
		logger.Log("event", "failed to update todo", "error", err)
		return nil, err
	}

	return td, nil
}

func (s *Service) DeleteTodo(ctx context.Context, db *gorm.DB, id uint) error {
	ctx, span := trace.StartSpan(ctx, "todo.delete")
	defer span.End()
	logger := logging.FromContext(ctx, s.Logger)

	err := db.Scopes(withTodoID(id)).Delete(&Todo{}).Error

	if err != nil {
		logger.Log("event", "failed to delete todo", "error", err)
		return err
	}

	return nil
}

func (s *Service) GetTodo(ctx context.Context, db *gorm.DB, id uint) (*Todo, error) {
	ctx, span := trace.StartSpan(ctx, "todo.get")
	defer span.End()
	logger := logging.FromContext(ctx, s.Logger)

	todos, err := findTodos(db, withTodoID(id))

	if err != nil {
		logger.Log("event", "failed to retrieve todo", "error", err)
		return nil, err
	}

	return &todos[0], nil
}

func (s *Service) GetTodos(ctx context.Context, db *gorm.DB, pg PaginateTodos) (*PaginatedTodos, error) {
	ctx, span := trace.StartSpan(ctx, "todo.list")
	defer span.End()
	logger := logging.FromContext(ctx, s.Logger)

	originalLimit := pg.Limit
	if originalLimit > DefaultMaxLimit {
		originalLimit = DefaultMaxLimit
	}
	if originalLimit == 0 {
		originalLimit = DefaultLimit
	}
	pg.Limit = originalLimit + 1

	var totalCount int
	scopes := buildPaginatedScope(pg)
	items, err := findTodos(db, scopes...)
	if err != nil {
		logger.Log("event", "failed to fetch todos", "error", err)
		return nil, err
	}

	scopes = append(scopes, database.WithOffset(0))
	err = db.Model(Todo{}).Scopes(scopes...).Count(&totalCount).Error
	if err != nil {
		return nil, err
	}

	hasMore := false
	if uint32(len(items)) > originalLimit {
		hasMore = true
		items = items[:len(items)-1]
	}

	pg.Limit = originalLimit
	return &PaginatedTodos{
		Items:      items,
		HasMore:    hasMore,
		TotalCount: totalCount,
	}, nil
}

func (s *Service) AddComment(ctx context.Context, db *gorm.DB, comment AddComment) (*Comment, error) {
	ctx, span := trace.StartSpan(ctx, "todo.comment.add")
	defer span.End()
	logger := logging.FromContext(ctx, s.Logger)

	cmnt := &Comment{
		Text:   comment.Text,
		TodoId: comment.TodoId,
	}

	err := db.Save(cmnt).Error
	if err != nil {
		logger.Log("event", "failed to store comment", "error", err)
		return nil, err
	}

	return cmnt, nil
}

func (s *Service) RemoveComment(ctx context.Context, db *gorm.DB, todoId, id uint) error {
	ctx, span := trace.StartSpan(ctx, "todo.comment.remove")
	defer span.End()
	logger := logging.FromContext(ctx, s.Logger)

	cmnt := &Comment{
		id,
		"",
		todoId,
	}

	err := db.Delete(cmnt).Error
	if err != nil {
		logger.Log("event", "failed to remove comment", "error", err)
		return err
	}

	return nil
}

func (s *Service) GetComments(ctx context.Context, db *gorm.DB, todoId uint) ([]Comment, error) {
	ctx, span := trace.StartSpan(ctx, "todo.comments.get")
	defer span.End()
	logger := logging.FromContext(ctx, s.Logger)

	todo := &Todo{todoId, "", time.Now(), false}

	comments := []Comment{}
	err := db.Model(todo).Related(&comments).Limit(MaxComments).Error

	if err != nil {
		logger.Log("event", "failed to retrieve comments", "error", err)
		return nil, err
	}

	return comments, nil
}

func (s *Service) AddLabel(ctx context.Context, db *gorm.DB, label AddLabel) (*Label, error) {
	ctx, span := trace.StartSpan(ctx, "todo.label.add")
	defer span.End()
	logger := logging.FromContext(ctx, s.Logger)

	lbl := &Label{
		TodoId: label.TodoId,
		Text:   label.Text,
		Color:  label.Color,
	}

	err := db.Create(lbl).Error
	if err != nil {
		logger.Log("event", "failed to store label", "error", err)
		return nil, err
	}

	return lbl, nil
}

func (s *Service) RemoveLabel(ctx context.Context, db *gorm.DB, todoId, id uint) error {
	ctx, span := trace.StartSpan(ctx, "todo.label.remove")
	defer span.End()
	logger := logging.FromContext(ctx, s.Logger)

	lbl := &Label{id,
		"",
		"",
		todoId,
	}

	err := db.Delete(lbl).Error
	if err != nil {
		logger.Log("event", "failed to remove label", "error", err)
		return err
	}

	return nil
}

func (s *Service) GetLabels(ctx context.Context, db *gorm.DB, todoId uint) ([]Label, error) {
	ctx, span := trace.StartSpan(ctx, "todo.labels.get")
	defer span.End()
	logger := logging.FromContext(ctx, s.Logger)

	todo := &Todo{todoId, "", time.Now(), false}

	labels := []Label{}
	err := db.Model(todo).Related(&labels).Limit(MaxLabels).Error

	if err != nil {
		logger.Log("event", "failed to retrieve labels", "error", err)
		return nil, err
	}

	return labels, nil
}

func findTodos(db *gorm.DB, scopes ...database.Scope) ([]Todo, error) {
	todos := []Todo{}
	err := db.Scopes(scopes...).Find(&todos).Error

	return todos, err
}
