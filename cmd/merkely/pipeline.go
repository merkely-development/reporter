package main

import (
	"io"

	"github.com/spf13/cobra"
)

const pipelineDesc = `All Merkely pipelines operations.`

func newPipelineCmd(out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pipeline",
		Short: pipelineDesc,
		Long:  pipelineDesc,
	}

	// Add subcommands
	cmd.AddCommand(
		newDeclarePipelineCmd(out),
		newArtifactCmd(out),
	)

	return cmd
}
