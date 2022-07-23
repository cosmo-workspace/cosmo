package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/cosmo-workspace/cosmo/pkg/auth"
	"github.com/cosmo-workspace/cosmo/pkg/clog"

	"github.com/gorilla/securecookie"
)

// Manager manages local ports and each proxy servers
type Manager struct {
	Log *clog.Logger

	Insecure          bool
	TLSCertPath       string
	TLSPrivateKeyPath string

	ProxyBackendScheme       string
	ProxyGracefulShutdownDur time.Duration
	ProxyStartupCheckTimeout time.Duration

	User          string
	MaxAgeSeconds int

	Authorizer auth.Authorizer

	proxyStore localPortProxyStore
	lock       sync.Mutex

	authKey    []byte
	encryptKey []byte
}

type localPortProxyStore map[string]dynamicLocalPortProxyInfo

type dynamicLocalPortProxyInfo struct {
	targetPort int
	localPort  int
	shutdown   context.CancelFunc
	errCh      chan error
}

func (s localPortProxyStore) Add(name string, proxyData dynamicLocalPortProxyInfo) {
	s[name] = proxyData
}

func (s localPortProxyStore) Get(name string) (dynamicLocalPortProxyInfo, bool) {
	data, exist := s[name]

	return data, exist
}

func (m *Manager) setupProxyStore() {
	m.proxyStore = make(localPortProxyStore)
}

func (m *Manager) Initialize() (*Manager, error) {
	m.authKey = securecookie.GenerateRandomKey(64)
	if m.authKey == nil {
		return nil, errors.New("failed to generate random authKey")
	}
	m.encryptKey = securecookie.GenerateRandomKey(32)
	if m.encryptKey == nil {
		return nil, errors.New("failed to generate random encryptKey")
	}

	m.setupProxyStore()
	return m, nil
}

func (m *Manager) newProxyServer(name string, targetPort int) *ProxyServer {
	p := &ProxyServer{
		Log:               m.Log.WithName("proxy").WithName(name),
		User:              m.User,
		StaticFileDir:     "./public",
		MaxAgeSeconds:     m.MaxAgeSeconds,
		SessionName:       "cosmo-auth-proxy",
		RedirectPath:      "/cosmo-auth-proxy",
		Insecure:          m.Insecure,
		TLSCertPath:       m.TLSCertPath,
		TLSPrivateKeyPath: m.TLSPrivateKeyPath,
	}
	targetURL := &url.URL{
		Scheme: m.ProxyBackendScheme,
		Host:   fmt.Sprintf("localhost:%d", targetPort),
	}
	p.SetupReverseProxy(":0", targetURL)
	p.SetupSessionStore(m.authKey, m.encryptKey)
	p.SetupAuthorizer(m.Authorizer)
	return p
}

func (m *Manager) CreateNewProxy(ctx context.Context, name string, targetPort int) (int, error) {
	m.lock.Lock()
	defer m.lock.Unlock()

	log := m.Log.WithCaller()
	_, exist := m.proxyStore[name]
	if exist {
		return 0, fmt.Errorf("%s already exist", name)
	}

	proxyCtx, shutdown := context.WithCancel(context.Background())

	proxy := m.newProxyServer(name, targetPort)

	errCh := make(chan error)
	go func() {
		errCh <- proxy.Start(proxyCtx, m.ProxyGracefulShutdownDur)
	}()

	localPort := proxy.GetListenerPort()
WaitStartLoop:
	for {
		select {
		case err := <-errCh:
			shutdown()
			return 0, fmt.Errorf("failed to start server: %w", err)

		case <-ctx.Done():
			shutdown()
			return 0, errors.New("canceled")

		default:
			if localPort != 0 {
				break WaitStartLoop
			}
			localPort = proxy.GetListenerPort()
			time.Sleep(time.Second)
		}
	}

	proxyInfo := dynamicLocalPortProxyInfo{
		targetPort: targetPort,
		localPort:  localPort,
		shutdown:   shutdown,
		errCh:      errCh,
	}
	m.proxyStore.Add(name, proxyInfo)

	proxyStartupCheckCtx, cancel := context.WithTimeout(ctx, m.ProxyStartupCheckTimeout)
	defer cancel()

HealthCheckLoop:
	for {
		select {
		case <-proxyStartupCheckCtx.Done():
			shutdown()
			return localPort, errors.New("failed to pass healthcheck")
		default:
			err := m.doHealthCheck(proxyStartupCheckCtx, proxyInfo)
			if err == nil {
				break HealthCheckLoop
			}
			log.DebugAll().Info("healthcheck failed", "error", err)
			time.Sleep(time.Second)
		}
	}

	log.Info("successfully created new auth proxy", "targetPort", proxyInfo.targetPort, "localPort", proxyInfo.localPort)
	return localPort, nil
}

type RunningProxyInfo struct {
	PortName   string
	TargetPort int
	LocalPort  int
}

func (m *Manager) GetRunningProxies() (runningProxies []RunningProxyInfo) {

	runningProxies = make([]RunningProxyInfo, 0, len(m.proxyStore))
	for portName, proxy := range m.proxyStore {
		runningProxies = append(runningProxies, RunningProxyInfo{
			PortName:   portName,
			TargetPort: proxy.targetPort,
			LocalPort:  proxy.localPort,
		})
	}
	sort.Slice(runningProxies, func(i, j int) bool { return runningProxies[i].PortName < runningProxies[j].PortName })
	return runningProxies
}

func (m *Manager) GetRunningProxy(name string) (localPort int, targetPort int, exist bool) {
	proxy, exist := m.proxyStore.Get(name)
	if !exist {
		return 0, 0, false
	}
	return proxy.localPort, proxy.targetPort, exist
}

func (m *Manager) ShutdownProxy(ctx context.Context, name string) error {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.shutdownProxy(ctx, name)
}

func (m *Manager) shutdownProxy(ctx context.Context, name string) error {
	proxy, exist := m.proxyStore[name]
	if !exist {
		return fmt.Errorf("%s not found in proxy store", name)
	}
	delete(m.proxyStore, name)

	if proxy.shutdown == nil {
		return errors.New("shutdown func is nil")
	}

	proxy.shutdown()

	waiter := time.NewTimer(m.ProxyGracefulShutdownDur)
	defer waiter.Stop()
	for {
		select {
		case <-waiter.C:
			return errors.New("reached to shutdown timeout")
		case <-ctx.Done():
			return errors.New("cancel")
		case err := <-proxy.errCh:
			return ignoreErrServerClosed(err)
		}
	}
}

func (m *Manager) GC(ctx context.Context, runningProxyNameList []string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	log := m.Log.WithName("GC")

	proxyUseCounts := make(map[string]int)
	for storedProxyName := range m.proxyStore {
		count := 0
		for _, runningProxyName := range runningProxyNameList {
			if storedProxyName == runningProxyName {
				count++
			}
		}
		proxyUseCounts[storedProxyName] = count
	}
	log.Debug().Info("proxyUseCounts", "map", proxyUseCounts)

	wg := sync.WaitGroup{}
	for name, count := range proxyUseCounts {
		if count == 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				log.Info("shutdown unused proxy", "name", name)
				m.shutdownProxy(ctx, name)
			}()
		}
	}
	wg.Wait()
}

func (m *Manager) doHealthCheck(ctx context.Context, proxy dynamicLocalPortProxyInfo) error {
	proto := "https"
	if m.Insecure {
		proto = "http"
	}

	url := fmt.Sprintf("%s://localhost:%d/", proto, proxy.localPort)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	// ref. http.DefaultTransport
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, // disable tls verification
	}
	client := &http.Client{Transport: transport}

	resp, err := client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return err
	}

	if resp.StatusCode < http.StatusOK || http.StatusMultipleChoices <= resp.StatusCode {
		return fmt.Errorf("invalid status code: %d", resp.StatusCode)
	}
	return nil
}

func ignoreErrServerClosed(err error) error {
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
