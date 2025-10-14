package log

import (
	"fmt"
	"go.uber.org/zap"
	"time"
)

type Field interface{}

// toZapFields converts a slice of Field interfaces to a slice of zap.Field.
func toZapFields(fields []Field) []zap.Field {
	var zapFields []zap.Field
	for _, f := range fields {
		if zf, ok := f.(zap.Field); ok {
			zapFields = append(zapFields, zf)
		}
	}
	return zapFields
}

func String(key, val string) Field {
	return zap.String(key, val)
}

func Stringer(key string, val fmt.Stringer) Field {
	return zap.Stringer(key, val)
}

func Int(key string, val int) Field {
	return zap.Int(key, val)
}

func Int64(key string, val int64) Field {
	return zap.Int64(key, val)
}

func Uint(key string, val uint) Field {
	return zap.Uint(key, val)
}

func Uint64(key string, val uint64) Field {
	return zap.Uint64(key, val)
}

func Float64(key string, val float64) Field {
	return zap.Float64(key, val)
}

func Bool(key string, val bool) Field {
	return zap.Bool(key, val)
}

func ErrorField(val error) Field {
	return zap.Error(val)
}

func Stack(val string) Field {
	return zap.Stack(val)
}

func Any(key string, val interface{}) Field {
	return zap.Any(key, val)
}

func Time(key string, val time.Time) Field {
	return zap.Time(key, val)
}
