package logzzap

import (
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"sync"

	"go.uber.org/zap/zapcore"
)

type logzSender interface {
	io.Writer

	Sync() error
	Send(payload []byte) error
}

type LogzCore struct {
	zapcore.LevelEnabler

	logger  logzSender
	appName string
	env     string

	lock             *sync.RWMutex
	additionalFields map[string]any
}

type Option func(*LogzCore)

func WithAppName(name string) Option {
	return func(lc *LogzCore) {
		lc.appName = name
	}
}

func WithEnvironment(env string) Option {
	return func(lc *LogzCore) {
		lc.env = env
	}
}

// NewLogzCore creates a new core to transmit logs to logz.
// Logz token and other options should be set before creating a new core
func NewLogzCore(sender logzSender, minLevel zapcore.Level, options ...Option) *LogzCore {
	core := &LogzCore{
		LevelEnabler:     minLevel,
		logger:           sender,
		additionalFields: make(map[string]any),
		lock:             new(sync.RWMutex),
	}

	for _, option := range options {
		option(core)
	}

	return core
}

// With provides structure
func (c *LogzCore) With(fields []zapcore.Field) zapcore.Core {
	m := fieldsToMap(fields)

	c.lock.Lock()
	maps.Copy(c.additionalFields, m)
	c.lock.Unlock()

	return &LogzCore{
		LevelEnabler:     c.LevelEnabler,
		logger:           c.logger,
		additionalFields: c.additionalFields,
		appName:          c.appName,
		env:              c.env,
		lock:             c.lock,
	}
}

// Check determines if this should be sent to roll bar based on LevelEnabler
func (c *LogzCore) Check(entry zapcore.Entry, checkedEntry *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checkedEntry.AddCore(entry, c)
	}

	return checkedEntry
}

func (c *LogzCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	m := fieldsToMap(fields)
	m["message"] = entry.Message
	m["level"] = entry.Level.String()

	if entry.Caller.Defined {
		m["caller.file"] = entry.Caller.File
		m["caller.function"] = fmt.Sprintf("%s:%d", entry.Caller.Function, entry.Caller.Line)
	}

	if len(c.appName) > 0 {
		m["app"] = c.appName
	}

	if len(c.env) > 0 {
		m["environment"] = c.env
	}

	c.lock.Lock()
	maps.Copy(m, c.additionalFields)
	clear(c.additionalFields)
	c.lock.Unlock()

	blob, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshaling fields: %w", err)
	}

	if err := c.logger.Send(blob); err != nil {
		return fmt.Errorf("sending bytes: %w", err)
	}

	return nil
}

func (c *LogzCore) Sync() error {
	return c.logger.Sync()
}
