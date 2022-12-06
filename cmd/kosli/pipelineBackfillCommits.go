package main

import (
	"encoding/json"
	"fmt"
	"github.com/kosli-dev/cli/internal/gitview"
	"io"
	"net/http"

	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type pipelineBackfillArtifactCommitsOptions struct {
	srcRepoRoot string
	payload     ArtifactCommitsBackfillPayload
}

type ArtifactCommitsBackfillPayload struct {
	RepoUrl     string                    `json:"repo_url"`
	CommitsList []*gitview.ArtifactCommit `json:"git_commit_list"`
}

func newPipelineBackfillArtifactCommitsCmd(out io.Writer) *cobra.Command {
	o := new(pipelineBackfillArtifactCommitsOptions)
	cmd := &cobra.Command{
		Use:    "backfill-commits PIPELINE-NAME",
		Short:  "Collect and report the changelog of each artifact in a Kosli pipeline.",
		Hidden: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			if len(args) < 1 {
				return ErrorBeforePrintingUsage(cmd, "pipeline name argument is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(out, args)
		},
	}

	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", repoRootFlag)

	err := RequireFlags(cmd, []string{"repo-root"})
	if err != nil {
		log.Fatalf("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *pipelineBackfillArtifactCommitsOptions) run(out io.Writer, args []string) error {
	// Get all artifacts for a pipeline
	// find repo URL
	// for each artifact,
	// 1) find the commit of the previous artifact
	// 2) get the commit list
	// 3) send a backfill request
	var err error
	pipelineName := args[0]

	gitView, err := gitview.New(o.srcRepoRoot)
	if err != nil {
		return err
	}

	o.payload.RepoUrl, err = gitView.RepoUrl()
	if err != nil {
		return err
	}

	pageNumber := 0
	for {
		pageNumber += 1
		artifactsRaw, err := getPipelineArtifacts(pipelineName, pageNumber)
		if err != nil {
			return err
		}
		if len(artifactsRaw) == 0 {
			return nil
		}
		for _, artifactRaw := range artifactsRaw {
			evidenceMap := artifactRaw["evidence"].(map[string]interface{})
			artifactData := evidenceMap["artifact"].(map[string]interface{})
			gitCommit := artifactData["git_commit"].(string)
			artifactDigest := artifactData["sha256"].(string)
			_, _ = fmt.Fprintf(out, "Digest: %s -- git commit: %s \n", artifactDigest, gitCommit)

			previousCommitUrl := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s/previous_commit",
				global.Host, global.Owner, pipelineName, artifactDigest)

			response, err := requests.DoBasicAuthRequest([]byte{}, previousCommitUrl, "", global.ApiToken,
				global.MaxAPIRetries, http.MethodGet, map[string]string{}, log)
			if err != nil {
				return err
			}

			var previousCommitResponse map[string]interface{}
			err = json.Unmarshal([]byte(response.Body), &previousCommitResponse)
			if err != nil {
				return err
			}

			o.payload.CommitsList, err = gitView.ChangeLog(gitCommit, previousCommit(out, previousCommitResponse))
			if err != nil {
				return err
			}

			for _, commitData := range o.payload.CommitsList {
				_, _ = fmt.Fprintf(out, "	Commit sha1: %s\n", commitData.Sha1)
			}

			url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s/backfill_commits", global.Host, global.Owner, pipelineName, artifactDigest)
			_, err = requests.SendPayload(o.payload, url, "", global.ApiToken,
				global.MaxAPIRetries, global.DryRun, http.MethodPut, log)
			if err != nil {
				return err
			}
		}
	}
}

func previousCommit(out io.Writer, previousCommitResponse map[string]interface{}) string {
	previousCommit := ""
	if previousCommitResponse["previous_commit"] != nil {
		previousCommit = previousCommitResponse["previous_commit"].(string)
		_, _ = fmt.Fprintf(out, "Previous commit: %s\n", previousCommit)
	}
	return previousCommit
}

// getPipelineArtifacts returns artifacts from a pipeline
func getPipelineArtifacts(pipelineName string, pageNumber int) ([]map[string]interface{}, error) {
	var artifacts []map[string]interface{}
	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/?page=%d&per_page=%d",
		global.Host, global.Owner, pipelineName, pageNumber, 15)
	response, err := requests.SendPayload([]byte{}, url, "", global.ApiToken,
		global.MaxAPIRetries, false, http.MethodGet, log)
	if err != nil {
		return artifacts, err
	}

	err = json.Unmarshal([]byte(response.Body), &artifacts)
	if err != nil {
		return artifacts, err
	}
	return artifacts, nil
}
