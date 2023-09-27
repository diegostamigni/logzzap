package logzzap

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"go.uber.org/zap/zapcore"
)

type logzSender interface {
	io.Writer

	Sync() error
	Send(payload []byte) error
}

type buffer interface {
	io.ReadWriter

	Reset()
	Bytes() []byte
}

type LogzCore struct {
	zapcore.LevelEnabler

	coreFields map[string]any
	logger     logzSender
	encoder    *json.Encoder
	buffer     buffer
}

// NewLogzCore creates a new core to transmit logs to logz.
// Logz token and other options should be set before creating a new core
func NewLogzCore(sender logzSender, minLevel zapcore.Level) *LogzCore {
	buf := new(bytes.Buffer)
	return &LogzCore{
		LevelEnabler: minLevel,
		coreFields:   make(map[string]any),
		logger:       sender,
		encoder:      json.NewEncoder(buf),
		buffer:       buf,
	}
}

// With provides structure
func (c *LogzCore) With(fields []zapcore.Field) zapcore.Core {
	fieldMap := fieldsToMap(fields)
	for k, v := range fieldMap {
		c.coreFields[k] = v
	}

	return c
}

// Check determines if this should be sent to roll bar based on LevelEnabler
func (c *LogzCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checkedEntry.AddCore(entry, c)
	}

	return checkedEntry
}

func (c *LogzCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	fieldsMap := fieldsToMap(fields)
	fieldsMap["message"] = entry.Message
	fieldsMap["level"] = entry.Level.String()

	err := c.encoder.Encode(fieldsMap)
	if err != nil {
		return fmt.Errorf("marshaling fields: %w", err)
	}

	if err := c.logger.Send(c.buffer.Bytes()); err != nil {
		return fmt.Errorf("sending bytes: %w", err)
	}

	c.buffer.Reset()
	return nil
}

func (c *LogzCore) Sync() error {
	return c.logger.Sync()
}
