package main

import (
	"fmt"
	"testing"

	log "github.com/kosli-dev/cli/internal/logger"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/kosli-dev/cli/internal/testHelpers"
	"github.com/stretchr/testify/suite"
)

// Define the suite, and absorb the built-in basic suite
// functionality from testify - including a T() method which
// returns the current testing context
type CommitEvidencePRGitlabCommandTestSuite struct {
	suite.Suite
	defaultKosliArguments string
	pipelineName          string
}

func (suite *CommitEvidencePRGitlabCommandTestSuite) SetupTest() {
	testHelpers.SkipIfEnvVarUnset(suite.T(), []string{"KOSLI_GITLAB_TOKEN"})

	suite.pipelineName = "gitlab-pr"
	global = &GlobalOpts{
		ApiToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpZCI6ImNkNzg4OTg5In0.e8i_lA_QrEhFncb05Xw6E_tkCHU9QfcY4OLTVUCHffY",
		Owner:    "docs-cmd-test-user",
		Host:     "http://localhost:8001",
	}
	suite.defaultKosliArguments = fmt.Sprintf(" --host %s --owner %s --api-token %s", global.Host, global.Owner, global.ApiToken)
	logger = log.NewStandardLogger()
	kosliClient = requests.NewKosliClient(1, false, logger)

	CreatePipeline(suite.pipelineName, suite.T())
}

func (suite *CommitEvidencePRGitlabCommandTestSuite) TestCommitEvidencePRGitlabCmd() {
	tests := []cmdTestCase{
		{
			name: "report Gitlab PR evidence works when no merge requests are found",
			cmd: `commit report evidence gitlab-mergerequest --name gl-pr --pipelines ` + suite.pipelineName + `
			          --build-url example.com --gitlab-org ewelinawilkosz  --repository merkely-gitlab-demo --commit 2ec23dda01fc85e3f94a2b5ea8cb8cf7e79c4ed6` + suite.defaultKosliArguments,
			golden: "no merge requests found for given commit: 2ec23dda01fc85e3f94a2b5ea8cb8cf7e79c4ed6\n" +
				"gitlab merge request evidence is reported to commit: 2ec23dda01fc85e3f94a2b5ea8cb8cf7e79c4ed6\n",
		},
		{
			name: "report Gitlab PR evidence works when there are merge requests",
			cmd: `commit report evidence gitlab-mergerequest --name gl-pr --pipelines ` + suite.pipelineName + `
			          --build-url example.com --gitlab-org ewelinawilkosz  --repository merkely-gitlab-demo --commit e6510880aecdc05d79104d937e1adb572bd91911` + suite.defaultKosliArguments,
			golden: "gitlab merge request evidence is reported to commit: e6510880aecdc05d79104d937e1adb572bd91911\n",
		},
		{
			wantError: true,
			name:      "report Gitlab PR evidence fails when --name is missing",
			cmd: `commit report evidence gitlab-mergerequest --pipelines ` + suite.pipelineName + `
			          --build-url example.com --gitlab-org ewelinawilkosz  --repository merkely-gitlab-demo --commit e6510880aecdc05d79104d937e1adb572bd91911` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"name\" not set\n",
		},
		{
			wantError: true,
			name:      "report Gitlab PR evidence fails when --owner is missing",
			cmd: `commit report evidence gitlab-mergerequest --name gl-pr --pipelines ` + suite.pipelineName + `
			          --build-url example.com --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6 --api-token foo --host bar`,
			golden: "Error: --owner is not set\n" +
				"Usage: kosli commit report evidence gitlab-mergerequest [flags]\n",
		},
		{
			wantError: true,
			name:      "report Gitlab PR evidence fails when --gitlab-org is missing",
			cmd: `commit report evidence gitlab-mergerequest --name gl-pr --pipelines ` + suite.pipelineName + `
			          --build-url example.com --repository cli --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"gitlab-org\" not set\n",
		},
		{
			wantError: true,
			name:      "report Gitlab PR evidence fails when --repository is missing",
			cmd: `commit report evidence gitlab-mergerequest --name gl-pr --pipelines ` + suite.pipelineName + `
			          --build-url example.com --gitlab-org kosli-dev --commit 73d7fee2f31ade8e1a9c456c324255212c30c2a6` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"repository\" not set\n",
		},
		{
			wantError: true,
			name:      "report Gitlab PR evidence fails when --commit is missing",
			cmd: `commit report evidence gitlab-mergerequest --name gl-pr --pipelines ` + suite.pipelineName + `
			          --build-url example.com --gitlab-org kosli-dev --repository cli` + suite.defaultKosliArguments,
			golden: "Error: required flag(s) \"commit\" not set\n",
		},
		{
			wantError: true,
			name:      "report Gitlab PR evidence fails when commit does not exist",
			cmd: `commit report evidence gitlab-mergerequest --name gl-pr --pipelines ` + suite.pipelineName + `
			          --build-url example.com --gitlab-org ewelinawilkosz --repository merkely-gitlab-demo --commit 73d7fee2f31ade8e1a9c456c324255212c3123ab` + suite.defaultKosliArguments,
			golden: "Error: GET https://gitlab.com/api/v4/projects/ewelinawilkosz/merkely-gitlab-demo/repository/commits/73d7fee2f31ade8e1a9c456c324255212c3123ab/merge_requests: 404 {message: 404 Commit Not Found}\n",
		},
		{
			wantError: true,
			name:      "report Gitlab PR evidence fails when --assert is used and commit has no PRs",
			cmd: `commit report evidence gitlab-mergerequest --name gl-pr --pipelines ` + suite.pipelineName + `
					  --assert
			          --build-url example.com --gitlab-org ewelinawilkosz --repository merkely-gitlab-demo --commit 2ec23dda01fc85e3f94a2b5ea8cb8cf7e79c4ed6` + suite.defaultKosliArguments,
			golden: "Error: no merge requests found for the given commit: 2ec23dda01fc85e3f94a2b5ea8cb8cf7e79c4ed6\n",
		},
		{
			wantError: true,
			name:      "report Gitlab PR evidence fails when --user-data is not found",
			cmd: `commit report evidence gitlab-mergerequest --name gl-pr --pipelines ` + suite.pipelineName + `
					  --user-data non-existing.json
			          --build-url example.com --gitlab-org ewelinawilkosz --repository merkely-gitlab-demo --commit e6510880aecdc05d79104d937e1adb572bd91911` + suite.defaultKosliArguments,
			golden: "Error: open non-existing.json: no such file or directory\n",
		},
	}

	runTestCmd(suite.T(), tests)
}

// In order for 'go test' to run this suite, we need to create
// a normal test function and pass our suite to suite.Run
func TestCommitEvidencePRGitlabCommandTestSuite(t *testing.T) {
	suite.Run(t, new(CommitEvidencePRGitlabCommandTestSuite))
}
