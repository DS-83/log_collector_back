package models

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type EventLevel int

const (
	LevelDebug EventLevel = 10
	LevelInfo  EventLevel = 20
	LevelWarn  EventLevel = 30
	LevelError EventLevel = 40
)

var levelByName = map[string]EventLevel{
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
}

func ParseEventLevel(s string) (EventLevel, bool) {
	l, ok := levelByName[strings.ToLower(s)]
	return l, ok
}

func (l EventLevel) String() string {
	for name, v := range levelByName {
		if v == l {
			return name
		}
	}
	return "unknown"
}

type Event struct {
	ID         int64
	EventID    uuid.UUID // генерирует клиент, для идемпотентности
	OccurredAt time.Time // когда событие случилось в приложении
	ReceivedAt time.Time // когда пришло в коллектор
	Source     string    // сервис/модуль-источник: 'api', 'worker', ...
	EventType  string    // 'auth.login', 'payment.failed', ...
	Level      EventLevel
	ActorID    *string        // id пользователя в основном приложении (без FK)
	RequestID  *string        // корреляция с запросом / trace
	Payload    map[string]any // nil трактуем как {}
}

type EventType struct {
	Type        string // 'auth.login', 'payment.failed', ...
	Description string
	Enabled     bool
}

type EventFilter struct {
	Query     string
	From      time.Time // нулевое значение = не задано
	To        time.Time // нулевое значение = не задано
	Limit     int
	Offset    int
	SortField string
	Order     string
}
