package cli

import (
	"fmt"

	"github.com/getlode/lode/internal/repo"
	"github.com/spf13/cobra"
)

func newRemoteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remote",
		Short: "Manage remotes in .dvc/config",
	}
	cmd.AddCommand(newRemoteAddCmd(), newRemoteModifyCmd())
	return cmd
}

func newRemoteAddCmd() *cobra.Command {
	var makeDefault bool
	cmd := &cobra.Command{
		Use:   "add <name> <url>",
		Short: "Add a remote (url s3://bucket/prefix)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := requireRepo()
			if err != nil {
				return err
			}
			if err := repo.SetRemote(r.ConfigPath(), repo.Remote{Name: args[0], URL: args[1]}, makeDefault); err != nil {
				return err
			}
			infof("remote %q added", args[0])
			return nil
		},
	}
	cmd.Flags().BoolVarP(&makeDefault, "default", "d", false, "set as the default remote")
	return cmd
}

func newRemoteModifyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "modify <name> <option> <value>",
		Short: "Modify a remote option (endpointurl, region, access_key_id, ...)",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := requireRepo()
			if err != nil {
				return err
			}
			cfg, err := repo.LoadConfig(r.ConfigPath())
			if err != nil {
				return err
			}
			rm := cfg.Remotes[args[0]]
			rm.Name = args[0]
			if err := setRemoteOption(&rm, args[1], args[2]); err != nil {
				return err
			}
			if err := repo.SetRemote(r.ConfigPath(), rm, false); err != nil {
				return err
			}
			infof("remote %q: %s = %s", args[0], args[1], args[2])
			return nil
		},
	}
}

func setRemoteOption(r *repo.Remote, option, value string) error {
	switch option {
	case "url":
		r.URL = value
	case "endpointurl":
		r.EndpointURL = value
	case "region":
		r.Region = value
	case "access_key_id":
		r.AccessKeyID = value
	case "secret_access_key":
		r.SecretAccessKey = value
	case "session_token":
		r.SessionToken = value
	case "profile":
		r.Profile = value
	default:
		return fmt.Errorf("unknown remote option: %s", option)
	}
	return nil
}
