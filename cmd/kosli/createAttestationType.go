package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

const createAttestationTypeShortDesc = `Create or update a Kosli attestation type.`

const createAttestationTypeLongDesc = createAttestationTypeShortDesc + ``

const createAttestationTypeExample = ` `

type createAttestationTypeOptions struct {
	payload        CreateAttestationTypePayload
	schemaFilePath string
	jqRules        []string
}

type JQEvaluatorPayload struct {
	ContentType string   `json:"content_type"`
	Rules       []string `json:"rules"`
}

func NewJQEvaluatorPayload(rules []string) JQEvaluatorPayload {
	return JQEvaluatorPayload{"jq", rules}
}

type CreateAttestationTypePayload struct {
	TypeName    string             `json:"name"`
	Description string             `json:"description"`
	Evaluator   JQEvaluatorPayload `json:"evaluator"`
}

func newCreateAttestationTypeCmd(out io.Writer) *cobra.Command {
	o := new(createAttestationTypeOptions)
	cmd := &cobra.Command{
		Use:     "attestation-type TYPE-NAME",
		Short:   createAttestationTypeShortDesc,
		Long:    createAttestationTypeLongDesc,
		Example: createAttestationTypeExample,
		Args:    cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Org", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", attestationTypeDescriptionFlag)
	cmd.Flags().StringVarP(&o.schemaFilePath, "schema", "s", "", attestationTypeSchemaFlag)
	cmd.Flags().StringArrayVar(&o.jqRules, "jq", []string{}, attestationTypeJqFlag)

	addDryRunFlag(cmd)
	return cmd
}

func (o *createAttestationTypeOptions) run(args []string) error {
	o.payload.TypeName = args[0]
	o.payload.Evaluator = NewJQEvaluatorPayload(o.jqRules)

	form, err := prepareAttestationTypeForm(o.payload, o.schemaFilePath)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v2/custom-attestation-types/%s", global.Host, global.Org)
	reqParams := &requests.RequestParams{
		Method: http.MethodPost,
		URL:    url,
		Form:   form,
		DryRun: global.DryRun,
		Token:  global.ApiToken,
	}
	_, err = kosliClient.Do(reqParams)
	if err == nil && !global.DryRun {
		logger.Info("attestation-type %s was created", o.payload.TypeName)
	}
	return err
}

func prepareAttestationTypeForm(payload interface{}, schemaFilePath string) ([]requests.FormItem, error) {
	form, err := newAttestationTypeForm(payload, schemaFilePath)
	if err != nil {
		return []requests.FormItem{}, err
	}
	return form, nil
}

// newAttestationTypeForm constructs a list of FormItems for an attestation-type
// form submission.
func newAttestationTypeForm(payload interface{}, schemaFilePath string) (
	[]requests.FormItem, error,
) {
	form := []requests.FormItem{
		{Type: "field", FieldName: "data_json", Content: payload},
	}

	if schemaFilePath != "" {
		form = append(form, requests.FormItem{Type: "file", FieldName: "type_schema", Content: schemaFilePath})
	}

	return form, nil
}
