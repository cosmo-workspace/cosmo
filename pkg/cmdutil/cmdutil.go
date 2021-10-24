package cmdutil

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cosmo-workspace/cosmo/pkg/clog"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

const (
	KustomizationFile = "kustomization.yaml"
)

func GetKubeConfig(path string) (*api.Config, error) {
	if path == "" {
		rule := clientcmd.NewDefaultClientConfigLoadingRules()
		return rule.Load()
	} else {
		return clientcmd.LoadFromFile(path)
	}
}

var inclusterNamespaceFile = "/var/run/secrets/kubernetes.io/serviceaccount/namespace"

func GetDefaultNamespace(cfg *api.Config, kubecontext string) string {
	if cfg == nil || len(cfg.Contexts) == 0 {
		b, _ := ioutil.ReadFile(inclusterNamespaceFile)
		if len(b) != 0 {
			return string(b)
		}
		return ""
	}
	var ctxName string
	if kubecontext == "" {
		ctxName = cfg.CurrentContext
	} else {
		ctxName = kubecontext
	}
	ctx, ok := cfg.Contexts[ctxName]
	if !ok {
		return ""
	}
	return ctx.Namespace
}

func KustomizeBuildCmd() ([]string, error) {
	kust, kustErr := exec.LookPath("kustomize")
	if kustErr != nil {
		kctl, kctlErr := exec.LookPath("kubectl")
		if kctlErr != nil {
			return nil, fmt.Errorf("kubectl nor kustomize found: kustmizr=%v, kubectl=%v", kustErr, kctlErr)
		}
		return []string{kctl, "kustomize"}, nil
	}
	return []string{kust, "build"}, nil
}

func ExecKustomize(ctx context.Context, dir string, kust *types.Kustomization) ([]byte, error) {
	log := clog.FromContext(ctx).WithCaller()

	kustomizeBuildCmd, err := KustomizeBuildCmd()
	if err != nil {
		return nil, err
	}
	log.Debug().Info("kustomize cmd", "cmd", kustomizeBuildCmd)

	kustYaml, err := yaml.Marshal(kust)
	if err != nil {
		return nil, err
	}
	log.Debug().Info(string(kustYaml), "obj", "kustomization.yaml")

	// create kustomization.yaml
	if err := CreateFile(dir, KustomizationFile, kustYaml); err != nil {
		return nil, err
	}
	defer RemoveFile(dir, KustomizationFile)

	// run kustomize build
	kustomizeCmd := append(kustomizeBuildCmd, dir)

	out, err := exec.CommandContext(ctx, kustomizeCmd[0], kustomizeCmd[1:]...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to exec kustomize : %w : %s", err, out)
	}
	return out, nil
}

func CreateFile(dir, fname string, data []byte) error {
	fullPath, err := filepath.Abs(dir + "/" + fname)
	if err != nil {
		return fmt.Errorf("invaid file path : %w", err)
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create %s : %w", fname, err)
	}
	defer f.Close()

	if _, err = f.Write(data); err != nil {
		return fmt.Errorf("failed to create %s : %w", fname, err)
	}
	return nil
}

func RemoveFile(dir, fname string) error {
	fullPath, err := filepath.Abs(dir + "/" + fname)
	if err != nil {
		return fmt.Errorf("invaid file path : %w", err)
	}
	return os.Remove(fullPath)
}

func PrintfColorErr(out io.Writer, msg string, a ...interface{}) {
	fmt.Fprintf(out, "\x1b[33m%s\x1b[0m", fmt.Sprintf(msg, a...))
}

func PrintfColorInfo(out io.Writer, msg string, a ...interface{}) {
	fmt.Fprintf(out, "\x1b[32m%s\x1b[0m", fmt.Sprintf(msg, a...))
}
