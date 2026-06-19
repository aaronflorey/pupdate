package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/aaronflorey/pupdate/internal/state"
	"github.com/spf13/cobra"
)

func newResetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset",
		Short: "Delete the local .pupdate state file",
		Long:  "Remove the .pupdate state file in the current directory so the next run treats all dependencies as stale.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("failed to resolve working directory: %w", err)
			}

			path := filepath.Join(dir, state.FileName)
			err = os.Remove(path)
			if err != nil && !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("failed to remove %s: %w", path, err)
			}

			if errors.Is(err, os.ErrNotExist) {
				fmt.Fprintf(cmd.OutOrStdout(), "no %s file found; nothing to reset\n", state.FileName)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "removed %s\n", path)
			}
			return nil
		},
	}

	return cmd
}
