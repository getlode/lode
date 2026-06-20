// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/getlode/lode/internal/repo"
	"github.com/spf13/cobra"
)

// isSecretOption reports whether a remote option holds a credential whose value
// must never be echoed.
func isSecretOption(option string) bool {
	switch option {
	case "secret_access_key", "session_token", "access_key_id":
		return true
	}
	return false
}

// optionValue returns the explicit value (args[2]) or, when omitted, reads it
// from stdin so secrets never land in argv / ps / shell history.
func optionValue(args []string) (string, error) {
	if len(args) == 3 {
		return args[2], nil
	}
	data, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && data == "" {
		return "", fmt.Errorf("no value given for %q and none on stdin", args[1])
	}
	return strings.TrimRight(data, "\r\n"), nil
}

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
		Use:   "modify <name> <option> [value]",
		Short: "Modify a remote option (endpointurl, region, access_key_id, ...)",
		Long: "Modify a remote option. Omit [value] to read it from stdin — recommended\n" +
			"for secrets so they never appear in argv (ps), shell history, or logs.\n" +
			"Example: printf '%s' \"$SECRET\" | lode remote modify r secret_access_key",
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			r, err := requireRepo()
			if err != nil {
				return err
			}
			name, option := args[0], args[1]
			value, err := optionValue(args)
			if err != nil {
				return err
			}
			cfg, err := repo.LoadConfig(r.ConfigPath())
			if err != nil {
				return err
			}
			rm := cfg.Remotes[name]
			rm.Name = name
			if err := setRemoteOption(&rm, option, value); err != nil {
				return err
			}
			if err := repo.SetRemote(r.ConfigPath(), rm, false); err != nil {
				return err
			}
			if isSecretOption(option) {
				infof("remote %q: %s set", name, option)
			} else {
				infof("remote %q: %s = %s", name, option, value)
			}
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
