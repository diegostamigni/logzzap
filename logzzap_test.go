package logzzap

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type mockSender struct {
}

func (m *mockSender) Write(b []byte) (int, error) {
	return 0, nil
}

func (m *mockSender) Sync() error {
	return nil
}

func (m *mockSender) Send(payload []byte) error {
	return nil
}

func newEntry(level zapcore.Level, msg ...string) zapcore.Entry {
	strBuilder := strings.Builder{}
	strBuilder.WriteString(fmt.Sprintf("Test %s Messsage", level))

	if len(msg) > 0 {
		strBuilder.WriteString(fmt.Sprintf(" %s", strings.Join(msg, " ")))
	}

	return zapcore.Entry{
		Level:      level,
		Time:       time.Now(),
		LoggerName: "test_logger",
		Message:    strBuilder.String(),
		Caller:     zapcore.EntryCaller{},
		Stack:      "No stack. Just test.",
	}
}

func newSender() (*mockSender, error) {
	return &mockSender{}, nil
}

func TestLogzCoreLevels(t *testing.T) {
	sender, err := newSender()
	if err != nil {
		t.Fatal(err)
	}

	rc := NewLogzCore(sender, zapcore.DebugLevel)
	if rc == nil {
		t.Fatalf("LogzCore not initialized")
	}
	if rc.LevelEnabler != zapcore.DebugLevel {
		t.Fatalf("rc.minLevel is not the correct value")
	}

	rc = NewLogzCore(sender, zapcore.ErrorLevel)
	if rc == nil {
		t.Fatalf("LogzCore not initialized")
	}

	if rc.LevelEnabler != zapcore.ErrorLevel {
		t.Fatalf("rc.minLevel is not the correct value")
	}
}

func TestNewLogzCore(t *testing.T) {
	sender, err := newSender()
	if err != nil {
		t.Fatal(err)
	}

	rc := NewLogzCore(sender, zapcore.ErrorLevel)
	if rc == nil {
		t.Fatalf("LogzCore not initialized")
	}

	if rc.LevelEnabler != zapcore.ErrorLevel {
		t.Fatalf("rc.minLevel is not the correct value")
	}

	coreFields := []zap.Field{zap.String("foo", "bar"), zap.String("moo", "cow")}
	rc.With(coreFields)
	if _, ok := rc.additionalFields["foo"]; !ok {
		t.Fatalf("core fields not stored")
	}

	if _, ok := rc.additionalFields["moo"]; !ok {
		t.Fatalf("core fields not stored")
	}

	debugEntry := newEntry(zapcore.DebugLevel)
	rc.Check(debugEntry, nil) // should do more here?
	if err := rc.Write(debugEntry, nil); err != nil {
		t.Errorf("writing debug message %v", err)
	}

	// all of these will write, there are no checks, but in a real scenario, they work. Probably need to revisit this test
	// Right now, this is just to make sure the different levels are entered into Logz correctly.
	if err := rc.Write(newEntry(zapcore.InfoLevel), nil); err != nil {
		t.Errorf("writing Info message %v", err)
	}
	if err := rc.Write(newEntry(zapcore.WarnLevel), nil); err != nil {
		t.Errorf("writing warn message %v", err)
	}
	if err := rc.Write(newEntry(zapcore.ErrorLevel), nil); err != nil {
		t.Errorf("writing error message %v", err)
	}
	if err := rc.Write(newEntry(zapcore.DPanicLevel), nil); err != nil {
		t.Errorf("writing dpanic message %v", err)
	}
	if err := rc.Write(newEntry(zapcore.PanicLevel), nil); err != nil {
		t.Errorf("writing panic message %v", err)
	}
	if err := rc.Write(newEntry(zapcore.FatalLevel), nil); err != nil {
		t.Errorf("writing fatal message %v", err)
	}

	if err := rc.Sync(); err != nil {
		t.Errorf("sync failed - %v", err)
	}
}
