package clog

import (
	"context"
	"fmt"
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
	LEVEL_INFO        = 0
	LEVEL_DEBUG       = 1
	LEVEL_DEBUG_ALL   = 2
	LEVEL_OBJECT_DUMP = 5
)

func IntoContext(ctx context.Context, logger *Logger) context.Context {
	return log.IntoContext(ctx, logger.Logger)
}

func FromContext(ctx context.Context) *Logger {
	if ctx != nil {
		logger := log.FromContext(ctx)
		return NewLogger(logger)
	}
	return NewLogger(logr.Discard())
}

func NewLogger(log logr.Logger) *Logger {
	return &Logger{Logger: log}
}

type Logger struct {
	logr.Logger
}

func (l *Logger) Enabled() bool {
	return l.Logger.Enabled()
}

func (l *Logger) WithName(name string) *Logger {
	return NewLogger(l.Logger.WithName(name))
}

func (l *Logger) WithValues(keysAndValues ...interface{}) *Logger {
	return NewLogger(l.Logger.WithValues(keysAndValues...))
}

func (l *Logger) WithCaller() *Logger {
	return l.WithName(caller())
}

func (l *Logger) Debug() *Logger {
	return NewLogger(l.Logger.V(LEVEL_DEBUG))
}

func (l *Logger) DebugAll() *Logger {
	return NewLogger(l.Logger.V(LEVEL_DEBUG_ALL))
}

func (l *Logger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.Logger.Error(err, msg, keysAndValues...)
}

func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.Logger.Info(msg, keysAndValues...)
}

type NamedObject interface {
	GetName() string
	DeepCopyObject() apiruntime.Object
}

func (l *Logger) DumpObject(scheme *apiruntime.Scheme, obj NamedObject, msg string) {
	if l.Logger.V(LEVEL_OBJECT_DUMP).Enabled() {
		debugObj := obj.DeepCopyObject()
		gvk, _ := apiutil.GVKForObject(debugObj, scheme)
		apiVersion := gvk.GroupVersion()
		kind := gvk.Kind
		name := obj.GetName()

		b, _ := yaml.Marshal(obj)
		l.Logger.V(LEVEL_OBJECT_DUMP).Info(fmt.Sprintf(`dump object: %s
------ %s
------ APIVersion: %s, Kind: %s, Name: %s
%s
------
`, msg, fileLine(), apiVersion, kind, name, b))
	}
}

func (l *Logger) PrintObjectDiff(x, y NamedObject) {
	if l.Logger.Enabled() {
		diff := cmp.Diff(x, y)
		if diff == "" {
			diff = "no difference"
		}
		l.Info(fmt.Sprintf("Object diff: %s", diff), "x", x.GetName(), "y", y.GetName())
	}
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
