package dashboard

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
)

// InClusterServer serves certificate for incluster workspace
// It implements https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/manager#Runnable
type InClusterServer struct {
	Log           *clog.Logger
	StaticFileDir string
	Port          int

	http *http.Server
}

type InClusterServeFiles struct {
	CAFile string
}

func copyFile(srcFile, dstFile string) error {
	data, err := os.ReadFile(srcFile)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dstFile, data, 0644); err != nil {
		return err
	}
	return nil
}

func NewInClusterServer(log *clog.Logger, port int, f InClusterServeFiles) (*InClusterServer, error) {
	tmp, err := os.MkdirTemp("", "cosmo-dashboard-incluster-server")
	if err != nil {
		return nil, fmt.Errorf("failed to create incluster server tmp dir: %w", err)
	}

	// copy files to tmp dir
	if err := copyFile(f.CAFile, filepath.Join(tmp, filepath.Base(cli.InClusterCAFile))); err != nil {
		return nil, fmt.Errorf("failed to copy file: src=%s dst=%s: %w", f.CAFile, filepath.Join(tmp, filepath.Base(cli.InClusterCAFile)), err)
	}

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	mux := http.NewServeMux()
	// setup serving static files
	mux.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir(tmp))))
	srv.Handler = (NewHTTPRequestLogger(log)).Middleware((mux))

	return &InClusterServer{
		Log:           log,
		StaticFileDir: tmp,
		Port:          port,
		http:          srv,
	}, nil
}

// Start run server
func (s *InClusterServer) Start(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		s.Log.Info("shutdown server")
		s.shutdown()
	}()

	s.Log.Info("start incluster server")
	return s.http.ListenAndServe()
}

func (s *InClusterServer) shutdown() error {
	gracefulShutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	return s.http.Shutdown(gracefulShutdownCtx)
}
