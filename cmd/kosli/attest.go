package main

import (
	"io"

	"github.com/spf13/cobra"
)

const attestDesc = `All Kosli attest commands.`

func newAttestCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attest",
		Short: attestDesc,
		Long:  attestDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newAttestArtifactCmd(out),
	)
	return cmd
}