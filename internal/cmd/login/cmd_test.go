package login

import (
	"bytes"
	"testing"

	"github.com/cosmo-workspace/cosmo/pkg/cli"
	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"
)

func TestCommandLogin(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cosmoctl login suite")
}

var _ = Describe("help", func() {
	It("should match snapshot", func() {
		cmd := &cobra.Command{}
		out := bytes.Buffer{}
		cmd.SetOut(&out)
		AddCommand(cmd, cli.NewRootOptions())
		cmd.SetArgs([]string{"login", "--help"})
		err := cmd.Execute()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out.String()).To(MatchSnapShot())
	})
})
