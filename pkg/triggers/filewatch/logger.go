package filewatch

import "github.com/Notifiarr/notifiarr/pkg/mnd"

/* All this just to give the tail module a custom logger. Its logger interface kinda sucks. */

type logger struct {
	Logger mnd.Logger
}

const (
	loggerPrefix = "File Watcher:"
	reqID        = "tailer"
)

/* Important ones! */

func (l *logger) Print(v ...any) {
	l.Logger.Print(reqID, l.pfx(v)...)
}

func (l *logger) Printf(format string, v ...any) {
	l.Logger.Printf(reqID, loggerPrefix+" "+format, v...)
}

func (l *logger) Println(v ...any) {
	l.Logger.Print(reqID, l.pfx(v)...)
}

/* Less important ones. */

func (l *logger) Fatal(v ...any) {
	l.Logger.Error(reqID, l.pfx(v)...)
}

func (l *logger) Fatalf(format string, v ...any) {
	l.Logger.Errorf(reqID, loggerPrefix+" "+format, v...)
}

func (l *logger) Fatalln(v ...any) {
	l.Logger.Error(reqID, l.pfx(v)...)
}

func (l *logger) Panic(v ...any) {
	l.Logger.Error(reqID, l.pfx(v)...)
}

func (l *logger) Panicf(format string, v ...any) {
	l.Logger.Errorf(reqID, loggerPrefix+" "+format, v...)
}

func (l *logger) Panicln(v ...any) {
	l.Logger.Error(reqID, l.pfx(v)...)
}

func (l *logger) pfx(v []any) []any {
	return append([]any{loggerPrefix}, v...)
}
