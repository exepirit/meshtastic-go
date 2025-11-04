package log

type (
	Logger interface {
		Debug(msg string, args ...any)
		Info(msg string, args ...any)
		Warn(msg string, args ...any)
		Error(msg string, args ...any)
	}
	NOOPLogger struct{}
)

func (NOOPLogger) Debug(msg string, args ...any) {
}

func (NOOPLogger) Info(msg string, args ...any) {
}

func (NOOPLogger) Warn(msg string, args ...any) {
}

func (NOOPLogger) Error(msg string, args ...any) {
}
