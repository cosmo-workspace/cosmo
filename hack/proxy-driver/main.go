package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cosmo-workspace/cosmo/internal/authproxy/proxy"
	"github.com/cosmo-workspace/cosmo/pkg/auth"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"github.com/gorilla/securecookie"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	var proxyPort int
	flag.IntVar(&proxyPort, "port", 9999, "proxy listen port")
	var targetPort int
	flag.IntVar(&targetPort, "target-port", 8080, "target port")
	var caCertPath string
	flag.StringVar(&caCertPath, "cacert", "ca.crt", "ca cert file path")
	var authURL string
	flag.StringVar(&authURL, "auth-url", "http://localhost:8443/api/v1alpha1/auth/login", "auth url")
	var user string
	flag.StringVar(&user, "user", "auth-proxy-test", "user")
	var insecure bool
	flag.BoolVar(&insecure, "insecure", true, "insecure")
	var authUiPath string
	flag.StringVar(&authUiPath, "auth-ui", "../../web/auth-proxy-ui/build_test", "auth ui file path")

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	logr := zap.New(zap.UseFlagOptions(&opts))

	f, err := os.Open(caCertPath)
	if err != nil {
		logr.Error(err, "failed to open CA cert file")
		os.Exit(1)
	}
	defer f.Close()

	caCert, err := io.ReadAll(f)
	if err != nil {
		logr.Error(err, "failed to read CA cert file")
		os.Exit(1)
	}
	f.Close()

	srv := &proxy.ProxyServer{
		Log:               clog.NewLogger(logr.WithName("proxy")),
		User:              user,
		StaticFileDir:     authUiPath,
		MaxAgeSeconds:     60,
		SessionName:       "proxy-driver-test",
		RedirectPath:      "/proxy-driver-test",
		Insecure:          insecure,
		TLSCertPath:       "./tls.crt",
		TLSPrivateKeyPath: "./tls.key",
	}
	targetURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("localhost:%d", targetPort),
	}
	srv.SetupReverseProxy(fmt.Sprintf(":%d", proxyPort), targetURL)

	authKey := securecookie.GenerateRandomKey(64)
	if authKey == nil {
		logr.Info("failed to generate random authKey")
		os.Exit(1)
	}
	encryptKey := securecookie.GenerateRandomKey(32)
	if encryptKey == nil {
		logr.Info("failed to generate random encryptKey")
		os.Exit(1)
	}

	srv.SetupSessionStore(authKey, encryptKey)

	a := auth.NewHTTPAuthorizer(authURL, caCert)
	srv.SetupAuthorizer(a)

	errChan := make(chan error)
	go func() {
		errChan <- srv.Start(ctx, time.Second)
	}()

	localPort := srv.GetListenerPort()
WaitStartLoop:
	for {
		select {
		case <-errChan:
			logr.Error(err, "failed to start server")
			os.Exit(1)
		default:
			switch {
			case localPort != 0:
				break WaitStartLoop
			default:
				localPort = srv.GetListenerPort()
				time.Sleep(time.Second)
			}
		}
	}

	logr.Info("proxy info", "localPort", localPort, "targetPort", targetPort)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)

	<-sig
	logr.Info("shutdown")
	cancel()

	if err := <-errChan; err != nil {
		logr.Error(err, "exit on error")
	}
}
