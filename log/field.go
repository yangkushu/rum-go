package log

import (
	"fmt"
	"github.com/yangkushu/rum-go/iface"
	"go.uber.org/zap"
	"time"
)

// toZapFields converts a slice of Field interfaces to a slice of zap.Field.
func toZapFields(fields []iface.Field) []zap.Field {
	var zapFields []zap.Field
	for _, f := range fields {
		if zf, ok := f.(zap.Field); ok {
			zapFields = append(zapFields, zf)
		}
	}
	return zapFields
}

func String(key, val string) iface.Field {
	return zap.String(key, val)
}

func Stringer(key string, val fmt.Stringer) iface.Field {
	return zap.Stringer(key, val)
}

func Int(key string, val int) iface.Field {
	return zap.Int(key, val)
}

func Int64(key string, val int64) iface.Field {
	return zap.Int64(key, val)
}

func Uint(key string, val uint) iface.Field {
	return zap.Uint(key, val)
}

func Uint64(key string, val uint64) iface.Field {
	return zap.Uint64(key, val)
}

func Float64(key string, val float64) iface.Field {
	return zap.Float64(key, val)
}

func Bool(key string, val bool) iface.Field {
	return zap.Bool(key, val)
}

func ErrorField(val error) iface.Field {
	return zap.Error(val)
}

func Stack(val string) iface.Field {
	return zap.Stack(val)
}

func Any(key string, val interface{}) iface.Field {
	return zap.Any(key, val)
}

func Time(key string, val time.Time) iface.Field {
	return zap.Time(key, val)
}
