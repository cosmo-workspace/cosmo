package version

import (
	"bytes"
	"testing"

	"github.com/cosmo-workspace/cosmo/pkg/cli"
	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"
)

func TestCommandVersion(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cosmoctl version suite")
}

var _ = Describe("help", func() {
	It("should match snapshot", func() {
		cmd := &cobra.Command{}
		out := bytes.Buffer{}
		cmd.SetOut(&out)
		AddCommand(cmd, cli.NewRootOptions())
		cmd.SetArgs([]string{"version", "--help"})
		err := cmd.Execute()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out.String()).To(MatchSnapShot())
	})
})

var _ = Describe("version", func() {
	It("should match snapshot", func() {
		cmd := &cobra.Command{}
		out := bytes.Buffer{}
		cmd.SetOut(&out)
		o := cli.NewRootOptions()
		o.Versions = cli.VersionInfo{
			Version: "v1.2.3",
			Commit:  "commitid",
			Date:    "2022-01-01",
		}
		AddCommand(cmd, o)
		cmd.SetArgs([]string{"version"})
		err := cmd.Execute()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out.String()).To(MatchSnapShot())
	})
})
