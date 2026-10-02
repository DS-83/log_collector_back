package http

import (
	e "log_collect/err"
	"log_collect/internal"
	"log_collect/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HandlerEvent struct {
	eventUc internal.EventUseCase
}

func NewHandlerEvent(uc internal.EventUseCase) *HandlerEvent {
	return &HandlerEvent{
		eventUc: uc,
	}
}

type event struct {
	ID         int64          `json:"id"`
	EventID    uuid.UUID      `json:"event_id"`
	OccurredAt time.Time      `json:"occurred_at"`
	ReceivedAt time.Time      `json:"received_at"`
	Source     string         `json:"source"`
	EventType  string         `json:"event_type"`
	Level      string         `json:"level"`
	ActorID    *string        `json:"actor_id"`
	RequestID  *string        `json:"request_id"`
	Payload    map[string]any `json:"payload"`
}

func eventFromModel(m *models.Event) *event {
	return &event{
		ID:         m.ID,
		EventID:    m.EventID,
		OccurredAt: m.OccurredAt,
		ReceivedAt: m.ReceivedAt,
		Source:     m.Source,
		EventType:  m.EventType,
		Level:      m.Level.String(),
		ActorID:    m.ActorID,
		RequestID:  m.RequestID,
		Payload:    m.Payload,
	}
}

type createEventInput struct {
	EventID    uuid.UUID      `json:"event_id"    binding:"required"`
	OccurredAt time.Time      `json:"occurred_at" binding:"required"`
	EventType  string         `json:"event_type"  binding:"required"`
	Level      string         `json:"level"       binding:"required"`
	ActorID    *string        `json:"actor_id"`
	RequestID  *string        `json:"request_id"`
	Payload    map[string]any `json:"payload"`
}

func (in *createEventInput) toModel() (*models.Event, error) {
	level, ok := models.ParseEventLevel(in.Level)
	if !ok {
		return nil, e.ErrInvalidReqData
	}

	return &models.Event{
		EventID:    in.EventID,
		OccurredAt: in.OccurredAt,
		EventType:  in.EventType,
		Level:      level,
		ActorID:    in.ActorID,
		RequestID:  in.RequestID,
		Payload:    in.Payload,
	}, nil
}

func (h *HandlerEvent) Create(c *gin.Context) {
	errMsg := "error create event"
	in := new(createEventInput)

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 64<<10) // 64 КБ

	if err := c.ShouldBindJSON(in); err != nil {
		handleBadRequest(c, err, errMsg)
		return
	}
	event, err := in.toModel()
	if err != nil {
		handleBadRequest(c, err, errMsg)
		return
	}
	// Get source form context
	key := ApiKeyFromContext(c)
	event.Source = key.Source

	err = h.eventUc.Create(c.Request.Context(), event)
	if err != nil {
		abortWithError(c, err, errMsg)
		return
	}

	c.Status(http.StatusCreated)
}

func (h *HandlerEvent) GetByID(c *gin.Context) {
	errMsg := "error get event by id"
	idString := c.Param("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		handleBadRequest(c, err, errMsg)
		return
	}

	event, err := h.eventUc.GetByID(c.Request.Context(), int64(id))
	if err != nil {
		handleError(c, err, errMsg)
		return
	}

	c.JSON(http.StatusOK, eventFromModel(event))
}

type listResponse struct {
	Events []event `json:"events"`
	Count  int     `json:"count"`
}

func (h *HandlerEvent) List(c *gin.Context) {
	errMsg := "error list events"
	filter, err := getFilterParams(c)
	if err != nil {
		handleBadRequest(c, err, errMsg)
		return
	}

	events, count, err := h.eventUc.List(c.Request.Context(), filter)
	if err != nil {
		handleError(c, err, errMsg)
		return
	}

	out := []event{}
	for _, e := range events {
		out = append(out, *eventFromModel(&e))
	}

	c.JSON(http.StatusOK, listResponse{Events: out, Count: count})
}

func getFilterParams(c *gin.Context) (*models.EventFilter, error) {
	f := &models.EventFilter{
		Query:     c.Query("query"),
		SortField: c.Query("sortField"),
		Order:     c.Query("order"),
	}
	var err error
	if s := c.Query("limit"); s != "" {
		if f.Limit, err = strconv.Atoi(s); err != nil {
			return nil, err
		}
	}
	if s := c.Query("offset"); s != "" {
		if f.Offset, err = strconv.Atoi(s); err != nil {
			return nil, err
		}
	}
	if s := c.Query("from"); s != "" {
		if f.From, err = time.Parse(time.RFC3339, s); err != nil {
			return nil, err
		}
	}
	if s := c.Query("to"); s != "" {
		if f.To, err = time.Parse(time.RFC3339, s); err != nil {
			return nil, err
		}
	}
	return f, nil
}
