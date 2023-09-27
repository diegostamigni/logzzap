package logzzap

import "go.uber.org/zap/zapcore"

func extractError(fields []zapcore.Field) error {
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range fields {
		f.AddTo(enc)
	}

	var foundError error
	for _, f := range fields {
		if f.Type == zapcore.ErrorType {
			foundError = f.Interface.(error)
		}
	}
	return foundError
}

func fieldsToMap(fields []zapcore.Field) map[string]any {
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range fields {
		f.AddTo(enc)
	}

	m := make(map[string]any)
	for k, v := range enc.Fields {
		m[k] = v
	}
	return m
}
