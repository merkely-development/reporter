package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/merkely-development/reporter/internal/requests"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const statusDesc = `
Check the status of Merkely server.
`

type statusOptions struct {
	assert bool
}

func newStatusCmd(out io.Writer) *cobra.Command {
	o := new(statusOptions)
	cmd := &cobra.Command{
		Use:   "status",
		Short: statusDesc,
		Long:  statusDesc,
		Args:  NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out)
		},
	}

	cmd.Flags().BoolVar(&o.assert, "assert", false, "Exit with non-zero code if Merkely server is not responding.")

	return cmd
}

func (o *statusOptions) run(out io.Writer) error {
	url := fmt.Sprintf("%s/ready", global.Host)
	response, err := requests.DoBasicAuthRequest([]byte{}, url, "", "", global.MaxAPIRetries, http.MethodGet, map[string]string{}, logrus.New())
	if err != nil {
		if o.assert {
			return fmt.Errorf("Merkely server %s is unresponsive", global.Host)
		}
		fmt.Print("Down")
	} else {
		fmt.Print(response.Body)
	}
	return nil
}
