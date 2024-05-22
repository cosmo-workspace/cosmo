package cmd

import (
	"bytes"
	"testing"

	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/cosmo-workspace/cosmo/pkg/cli"
)

func TestCommandRoot(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cosmoctl root suite")
}

var _ = Describe("help", func() {
	It("should match snapshot", func() {
		o := cli.NewRootOptions()
		rootCmd := NewRootCmd(o)

		out := bytes.Buffer{}
		rootCmd.SetOut(&out)
		rootCmd.SetArgs([]string{"--help"})
		err := rootCmd.Execute()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out.String()).To(MatchSnapShot())
	})
})
