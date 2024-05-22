package template

import (
	"bytes"
	"testing"

	"github.com/cosmo-workspace/cosmo/pkg/cli"
	. "github.com/cosmo-workspace/cosmo/pkg/snap"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/spf13/cobra"
)

func TestCommandTemplate(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cosmoctl template suite")
}

var _ = Describe("help", func() {
	Context("template", func() {
		It("should match snapshot", func() {
			cmd := &cobra.Command{}
			out := bytes.Buffer{}
			cmd.SetOut(&out)
			AddCommand(cmd, cli.NewRootOptions())
			cmd.SetArgs([]string{"template", "--help"})
			err := cmd.Execute()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(out.String()).To(MatchSnapShot())
		})
	})

	Context("template generate", func() {
		It("should match snapshot", func() {
			cmd := &cobra.Command{}
			out := bytes.Buffer{}
			cmd.SetOut(&out)
			AddCommand(cmd, cli.NewRootOptions())
			cmd.SetArgs([]string{"template", "generate", "--help"})
			err := cmd.Execute()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(out.String()).To(MatchSnapShot())
		})
	})

	Context("template get", func() {
		It("should match snapshot", func() {
			cmd := &cobra.Command{}
			out := bytes.Buffer{}
			cmd.SetOut(&out)
			AddCommand(cmd, cli.NewRootOptions())
			cmd.SetArgs([]string{"template", "get", "--help"})
			err := cmd.Execute()
			Expect(err).ShouldNot(HaveOccurred())
			Expect(out.String()).To(MatchSnapShot())
		})
	})
})
