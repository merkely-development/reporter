package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/output"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const approvalLsDesc = `List a number of approvals in a pipeline.`

type approvalLsOptions struct {
	output     string
	pageNumber int
	pageLimit  int
}

func newApprovalLsCmd(out io.Writer) *cobra.Command {
	o := new(approvalLsOptions)
	cmd := &cobra.Command{
		Use:     "ls PIPELINE-NAME",
		Aliases: []string{"list"},
		Short:   approvalLsDesc,
		Long:    approvalLsDesc,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "pipeline name argument is required")
			}
			if o.pageNumber <= 0 {
				return ErrorBeforePrintingUsage(cmd, "page number must be a positive integer")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVarP(&o.output, "output", "o", "table", outputFlag)
	cmd.Flags().IntVar(&o.pageNumber, "page", 1, pageNumberFlag)
	cmd.Flags().IntVarP(&o.pageLimit, "page-limit", "n", 15, pageLimitFlag)

	return cmd
}

func (o *approvalLsOptions) run(out io.Writer, args []string) error {
	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/approvals/?page=%d&per_page=%d",
		global.Host, global.Owner, args[0], o.pageNumber, o.pageLimit)
	response, err := requests.SendPayload([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, false, http.MethodGet)
	if err != nil {
		return err
	}

	return output.FormattedPrint(response.Body, o.output, out, o.pageNumber,
		map[string]output.FormatOutputFunc{
			"table": printApprovalListAsTable,
			"json":  output.PrintJson,
		})

}

func printApprovalListAsTable(raw string, out io.Writer, page int) error {
	var approvals []map[string]interface{}
	err := json.Unmarshal([]byte(raw), &approvals)
	if err != nil {
		return err
	}

	if len(approvals) == 0 {
		msg := "No approvals were found"
		if page != 1 {
			msg = fmt.Sprintf("%s at page number %d", msg, page)
		}
		fmt.Fprintln(out, msg)
		return nil
	}

	header := []string{"ID", "ARTIFACT", "STATE", "LAST_MODIFIED_AT"}
	rows := []string{}
	for _, approval := range approvals {
		approvalId := int(approval["release_number"].(float64))
		artifactName := approval["artifact_name"].(string)
		approvalState := approval["state"].(string)
		artifactDigest := approval["base_artifact"].(string)
		lastModifiedAt, err := formattedTimestamp(approval["last_modified_at"], true)
		if err != nil {
			return err
		}
		row := fmt.Sprintf("%d\tName: %s\t%s\t%s", approvalId, artifactName, approvalState, lastModifiedAt)
		rows = append(rows, row)
		row = fmt.Sprintf("\tFingerprint: %s\t\t", artifactDigest)
		rows = append(rows, row)
		rows = append(rows, "\t\t\t")
	}
	tabFormattedPrint(out, header, rows)

	return nil
}
