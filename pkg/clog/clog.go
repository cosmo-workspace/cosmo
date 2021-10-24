package clog

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/yaml"
)

const (
	LEVEL_INFO      = 0
	LEVEL_DEBUG     = 1
	LEVEL_DEBUG_ALL = 2
)

func LogrIntoContext(ctx context.Context, logger logr.Logger) context.Context {
	return log.IntoContext(ctx, logger)
}

func IntoContext(ctx context.Context, logger *Logger) context.Context {
	return LogrIntoContext(ctx, logger.logger)
}

func FromContext(ctx context.Context) *Logger {
	if ctx != nil {
		logger := log.FromContext(ctx)
		return NewLogger(logger)
	}
	return NewLogger(log.NullLogger{})
}

func NewLogger(log logr.Logger) *Logger {
	return &Logger{logger: log, out: os.Stdout}
}

type Logger struct {
	logger logr.Logger
	out    io.Writer
}

func (l *Logger) Enabled() bool {
	return l.logger.Enabled()
}

func (l *Logger) WithName(name string) *Logger {
	return NewLogger(l.logger.WithName(name))
}

func (l *Logger) WithValues(keysAndValues ...interface{}) *Logger {
	return NewLogger(l.logger.WithValues(keysAndValues...))
}

func (l *Logger) WithCaller() *Logger {
	return l.WithName(caller())
}

func (l *Logger) Debug() *Logger {
	return NewLogger(l.logger.V(LEVEL_DEBUG))
}

func (l *Logger) DebugAll() *Logger {
	return NewLogger(l.logger.V(LEVEL_DEBUG_ALL))
}

func (l *Logger) Error(err error, msg string, keysAndValues ...interface{}) {
	if l.logger.Enabled() {
		l.logger.Error(err, msg, keysAndValues...)
	}
}

func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.logger.Info(msg, keysAndValues...)
}

type NamedObject interface {
	GetName() string
	DeepCopyObject() apiruntime.Object
}

func (l *Logger) DumpObject(scheme *apiruntime.Scheme, obj NamedObject, msg string) {
	if l.logger.Enabled() {
		debugObj := obj.DeepCopyObject()
		gvk, err := apiutil.GVKForObject(debugObj, scheme)
		if err != nil {
			return
		}
		apiVersion := gvk.GroupVersion()
		kind := gvk.Kind
		name := obj.GetName()

		b, err := yaml.Marshal(obj)
		if err == nil {
			fmt.Fprintf(l.out, `--- dump object: %s
--- %s
--- APIVersion: %s, Kind: %s, Name: %s
%s
`, msg, fileLine(), apiVersion, kind, name, b)
		}
	}
}

func (l *Logger) PrintObjectDiff(x, y interface{}) {
	if l.logger.Enabled() {
		PrintObjectDiff(l.out, x, y)
	}
}

func Diff(x, y interface{}) string {
	return cmp.Diff(x, y)
}

func PrintObjectDiff(out io.Writer, x, y interface{}) {
	diff := Diff(x, y)
	fmt.Fprintln(out, diff)
}

func caller() string {
	pc, _, _, ok := runtime.Caller(2)
	if ok {
		name := strings.Split(runtime.FuncForPC(pc).Name(), ".")
		if len(name) == 0 {
			return ""
		}
		return name[len(name)-1]
	}
	return ""
}

func fileLine() string {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		return fmt.Sprintf("%s:%d", file, line)
	}
	return ""
}
