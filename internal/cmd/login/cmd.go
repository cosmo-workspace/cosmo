package login

import (
	"context"
	"fmt"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/cosmo-workspace/cosmo/pkg/cli"
	"github.com/cosmo-workspace/cosmo/pkg/clog"
	dashv1alpha1 "github.com/cosmo-workspace/cosmo/proto/gen/dashboard/v1alpha1"
)

func AddCommand(cmd *cobra.Command, o *cli.RootOptions) {
	loginCmd := &cobra.Command{
		Use:   "login USER_NAME",
		Short: "Login to COSMO Dashboard Server",
		Example: `
  # interactive mode
  cosmoctl login

  # non interactive mode
  echo $PASSWORD | cosmoctl login USER_NAME --dashboard-url https://DASHBOARD_URL --password-stdin
`,
	}
	cmd.AddCommand(LoginCmd(loginCmd, o))
}

type LoginOption struct {
	*cli.RootOptions

	UserName      string
	Password      string
	PasswordStdin bool
	Again         bool
}

func LoginCmd(cmd *cobra.Command, opt *cli.RootOptions) *cobra.Command {
	o := &LoginOption{RootOptions: opt}
	cmd.RunE = cli.ConnectErrorHandler(o)
	cmd.Flags().BoolVar(&o.PasswordStdin, "password-stdin", false, "input new password from stdin pipe")
	cmd.Flags().BoolVar(&o.Again, "again", false, "login again")
	return cmd
}

func (o *LoginOption) Validate(cmd *cobra.Command, args []string) error {
	if o.UseKubeAPI {
		return fmt.Errorf("login command does not support using Kubernetes API")
	}
	if !o.Again && o.PasswordStdin {
		if o.DashboardURL == "" || len(args) == 0 {
			return fmt.Errorf("dashboard URL and user name are required by args when using --password-stdin")
		}
	}

	if err := o.RootOptions.Validate(cmd, args); err != nil {
		return err
	}
	return nil
}

func (o *LoginOption) Complete(cmd *cobra.Command, args []string) error {
	// check if config file already exists
	cfgPath, _ := o.GetConfigFilePath()
	previousLogin, _ := cli.NewOrLoadConfigFile(cfgPath)

	if o.Again {
		if previousLogin.Endpoint == "" {
			return fmt.Errorf("failed to get previous login state. please login without --again")
		}
		o.DashboardURL = previousLogin.Endpoint
		o.UserName = previousLogin.User

	} else {
		// 1. Ask Dashboard URL
		if o.DashboardURL == "" {
			prompt := "Dashboard URL: "
			if previousLogin.Endpoint != "" {
				prompt = fmt.Sprintf("Dashboard URL (%s): ", previousLogin.Endpoint)
			}
			input, err := cli.AskInput(prompt, false)
			if err != nil {
				return err
			}
			if input == "" {
				o.DashboardURL = previousLogin.Endpoint
			} else {
				o.DashboardURL = input
			}
		}

		// 2. Ask UserName
		if len(args) > 0 {
			o.UserName = args[0]
		}
		if o.UserName == "" {
			prompt := "User Name    : "
			if previousLogin.User != "" {
				prompt = fmt.Sprintf("User Name (%s): ", previousLogin.User)
			}
			input, err := cli.AskInput(prompt, false)
			if err != nil {
				return err
			}
			if input == "" {
				o.UserName = previousLogin.User
			} else {
				o.UserName = input
			}
		}
	}

	// 3. Ask Password
	if o.PasswordStdin {
		input, err := cli.ReadFromPipedStdin()
		if err != nil {
			return fmt.Errorf("failed to read from stdin pipe: %w", err)
		}
		o.Password = input
	} else {
		prompt := "Password     : "
		if previousLogin.Endpoint == "" {
			prompt = "Password: "
		}
		input, err := cli.AskInput(prompt, true)
		if err != nil {
			return err
		}
		o.Password = input
	}

	o.RootOptions.DisableUseServiceAccount = true
	if err := o.RootOptions.Complete(cmd, args); err != nil {
		return err
	}
	o.Logr.Debug().Info("input", "dashboardURL", o.DashboardURL, "userName", o.UserName, "password", mask(o.Password))

	cmd.SilenceErrors = true
	cmd.SilenceUsage = true
	return nil
}

func mask(s string) string {
	if s == "" {
		return ""
	}
	return "******"
}

func (o *LoginOption) RunE(cmd *cobra.Command, args []string) error {
	if err := o.Validate(cmd, args); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}
	if err := o.Complete(cmd, args); err != nil {
		return fmt.Errorf("invalid options: %w", err)
	}

	ctx, cancel := context.WithTimeout(o.Ctx, time.Second*10)
	defer cancel()
	ctx = clog.IntoContext(ctx, o.Logr)

	c := o.CosmoDashClient
	res, err := c.AuthServiceClient.Login(ctx, connect.NewRequest(&dashv1alpha1.LoginRequest{UserName: o.UserName, Password: o.Password}))
	if err != nil {
		return fmt.Errorf("failed to login: %w", err)
	}
	o.CliConfig.Token = cli.ExtractSessionToken(res)
	o.CliConfig.User = o.UserName
	o.CliConfig.Endpoint = o.GetDashboardURL()
	o.CliConfig.UseServiceAccount = false

	// reset cacert if endpoint is not in cluster
	if o.CliConfig.Endpoint != cli.InClusterDashboardURL {
		o.CliConfig.CACert = ""
	}

	// save session
	err = o.CliConfig.Save()
	if err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), color.GreenString("Successfully logined to %s as %s", o.CliConfig.Endpoint, o.CliConfig.User))

	return nil

}
