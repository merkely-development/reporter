package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/object"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/kosli-dev/cli/internal/requests"
	"github.com/spf13/cobra"
)

type artifactCreationOptions struct {
	fingerprintOptions *fingerprintOptions
	pipelineName       string
	srcRepoRoot        string
	payload            ArtifactPayload
}

type ArtifactPayload struct {
	Sha256      string            `json:"sha256"`
	Filename    string            `json:"filename"`
	Description string            `json:"description"`
	GitCommit   string            `json:"git_commit"`
	BuildUrl    string            `json:"build_url"`
	CommitUrl   string            `json:"commit_url"`
	RepoUrl     string            `json:"repo_url"`
	CommitsList []*ArtifactCommit `json:"commits_list"`
}

type ArtifactCommit struct {
	Sha1      string   `json:"sha1"`
	Message   string   `json:"message"`
	Author    string   `json:"author"`
	Timestamp int64    `json:"timestamp"`
	Branch    string   `json:"branch"`
	Parents   []string `json:"parents"`
}

const artifactCreationExample = `
# Report to a Kosli pipeline that a file type artifact has been created
kosli pipeline artifact report creation FILE.tgz \
	--api-token yourApiToken \
	--artifact-type file \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--owner yourOrgName \
	--pipeline yourPipelineName 

# Report to a Kosli pipeline that an artifact with a provided fingerprint (sha256) has been created
kosli pipeline artifact report creation \
	--api-token yourApiToken \
	--build-url https://exampleci.com \
	--commit-url https://github.com/YourOrg/YourProject/commit/yourCommitShaThatThisArtifactWasBuiltFrom \
	--git-commit yourCommitShaThatThisArtifactWasBuiltFrom \
	--owner yourOrgName \
	--pipeline yourPipelineName \
	--sha256 yourSha256 
`

//goland:noinspection GoUnusedParameter
func newArtifactCreationCmd(out io.Writer) *cobra.Command {
	o := new(artifactCreationOptions)
	o.fingerprintOptions = new(fingerprintOptions)
	cmd := &cobra.Command{
		Use:     "creation ARTIFACT-NAME-OR-PATH",
		Short:   "Report an artifact creation to a Kosli pipeline. ",
		Long:    artifactCreationDesc(),
		Example: artifactCreationExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := RequireGlobalFlags(global, []string{"Owner", "ApiToken"})
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}

			err = ValidateArtifactArg(args, o.fingerprintOptions.artifactType, o.payload.Sha256, true)
			if err != nil {
				return ErrorBeforePrintingUsage(cmd, err.Error())
			}
			return ValidateRegistryFlags(cmd, o.fingerprintOptions)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return o.run(args)
		},
	}

	ci := WhichCI()
	cmd.Flags().StringVarP(&o.payload.Sha256, "sha256", "s", "", sha256Flag)
	cmd.Flags().StringVarP(&o.pipelineName, "pipeline", "p", "", pipelineNameFlag)
	cmd.Flags().StringVarP(&o.payload.Description, "description", "d", "", artifactDescriptionFlag)
	cmd.Flags().StringVarP(&o.payload.GitCommit, "git-commit", "g", DefaultValue(ci, "git-commit"), gitCommitFlag)
	cmd.Flags().StringVarP(&o.payload.BuildUrl, "build-url", "b", DefaultValue(ci, "build-url"), buildUrlFlag)
	cmd.Flags().StringVarP(&o.payload.CommitUrl, "commit-url", "u", DefaultValue(ci, "commit-url"), commitUrlFlag)
	cmd.Flags().StringVar(&o.srcRepoRoot, "repo-root", ".", repoRootFlag)
	addFingerprintFlags(cmd, o.fingerprintOptions)

	err := RequireFlags(cmd, []string{"pipeline", "git-commit", "build-url", "commit-url"})
	if err != nil {
		logger.Error("failed to configure required flags: %v", err)
	}

	return cmd
}

func (o *artifactCreationOptions) run(args []string) error {
	if o.payload.Sha256 != "" {
		o.payload.Filename = args[0]
	} else {
		var err error
		o.payload.Sha256, err = GetSha256Digest(args[0], o.fingerprintOptions)
		if err != nil {
			return err
		}
		if o.fingerprintOptions.artifactType == "dir" || o.fingerprintOptions.artifactType == "file" {
			o.payload.Filename = filepath.Base(args[0])
		} else {
			o.payload.Filename = args[0]
		}
	}

	previousCommit, err := latestCommit(o.pipelineName, o.payload.Sha256)
	if err != nil {
		return err
	}

	o.payload.CommitsList, err = changeLog(o.srcRepoRoot, o.payload.GitCommit, previousCommit)
	if err != nil {
		return err
	}

	o.payload.RepoUrl, err = getRepoUrl(o.srcRepoRoot)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/", global.Host, global.Owner, o.pipelineName)
	_, err = requests.SendPayload(o.payload, url, "", global.ApiToken,
		global.MaxAPIRetries, global.DryRun, http.MethodPut)
	return err
}

// changeLog attempts to collect the changelog list of commits for an artifact,
// the changelog is all commits between current commit and the commit from which the previous artifact in Kosli
// was created.
// If collecting the changelog fails (e.g. if git history has been rewritten), the changelog only
// contains the single commit info which is the current commit
func changeLog(srcRepoRoot, currentCommit, previousCommit string) ([]*ArtifactCommit, error) {
	if previousCommit != "" {
		commitsList, err := listCommitsBetween(srcRepoRoot, previousCommit, currentCommit)
		if err != nil {
			fmt.Printf("Warning: %s\n", err)
		} else {
			return commitsList, nil
		}
	}

	currentArtifactCommit, err := newArtifactCommitFromGitCommit(srcRepoRoot, currentCommit)
	if err != nil {
		return []*ArtifactCommit{}, fmt.Errorf("could not retrieve current git commit for %s: %v", currentCommit, err)
	}
	return []*ArtifactCommit{currentArtifactCommit}, nil
}

// latestCommit retrieves the git commit of the latest artifact for a pipeline in Kosli
func latestCommit(pipelineName, fingerprint string) (string, error) {
	latestCommitUrl := fmt.Sprintf("%s/api/v1/projects/%s/%s/artifacts/%s/latest_commit",
		global.Host, global.Owner, pipelineName, fingerprint)

	response, err := requests.DoBasicAuthRequest([]byte{}, latestCommitUrl, "", global.ApiToken,
		global.MaxAPIRetries, http.MethodGet, map[string]string{})
	if err != nil {
		return "", err
	}

	var latestCommitResponse map[string]interface{}
	err = json.Unmarshal([]byte(response.Body), &latestCommitResponse)
	if err != nil {
		return "", err
	}
	latestCommit := latestCommitResponse["latest_commit"]
	if latestCommit == nil {
		return "", nil
	} else {
		return latestCommit.(string), nil
	}
}

// newArtifactCommitFromGitCommit returns an ArtifactCommit object from a git commit
// the gitCommit can be a revision: e.g. HEAD or HEAD~2 etc
func newArtifactCommitFromGitCommit(srcRepoRoot, gitCommit string) (*ArtifactCommit, error) {
	repo, err := git.PlainOpen(srcRepoRoot)
	if err != nil {
		return &ArtifactCommit{}, fmt.Errorf("failed to open git repository at %s: %v", srcRepoRoot, err)
	}

	branchName, err := branchName(repo)
	if err != nil {
		return &ArtifactCommit{}, err
	}

	currentHash, err := repo.ResolveRevision(plumbing.Revision(gitCommit))
	if err != nil {
		return &ArtifactCommit{}, fmt.Errorf("failed to resolve %s: %v", gitCommit, err)
	}
	currentCommit, err := repo.CommitObject(*currentHash)
	if err != nil {
		return &ArtifactCommit{}, fmt.Errorf("could not retrieve commit for %s: %v", *currentHash, err)
	}

	return asArtifactCommit(currentCommit, branchName), nil
}

// getRepoUrl returns HTTPS URL for the `origin` remote of a repo
func getRepoUrl(repoRoot string) (string, error) {
	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return "", fmt.Errorf("failed to open git repository at %s: %v",
			repoRoot, err)
	}
	repoRemote, err := repo.Remote("origin") // TODO: We hard code this for now. Should we have a flag to set it from the cmdline?
	if err != nil {
		fmt.Printf("Warning: Repo URL will not be reported since there is no remote('origin') in git repository (%s)\n", repoRoot)
		return "", nil
	}
	remoteUrl := repoRemote.Config().URLs[0]
	if strings.HasPrefix(remoteUrl, "git@") {
		remoteUrl = strings.Replace(remoteUrl, ":", "/", 1)
		remoteUrl = strings.Replace(remoteUrl, "git@", "https://", 1)
	}
	remoteUrl = strings.TrimSuffix(remoteUrl, ".git")
	return remoteUrl, nil
}

// listCommitsBetween list all commits that have happened between two commits in a git repo
func listCommitsBetween(repoRoot, oldest, newest string) ([]*ArtifactCommit, error) {
	// Using 'var commits []*ArtifactCommit' will make '[]' convert to 'null' when converting to json
	// which will fail on the server side.
	// Using 'commits := make([]*ArtifactCommit, 0)' will make '[]' convert to '[]' when converting to json
	// See issue #522
	commits := make([]*ArtifactCommit, 0)
	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return commits, fmt.Errorf("failed to open git repository at %s: %v",
			repoRoot, err)
	}

	branchName, err := branchName(repo)
	if err != nil {
		return commits, err
	}

	newestHash, err := repo.ResolveRevision(plumbing.Revision(newest))
	if err != nil {
		return commits, fmt.Errorf("failed to resolve %s: %v", newest, err)
	}
	oldestHash, err := repo.ResolveRevision(plumbing.Revision(oldest))
	if err != nil {
		return commits, fmt.Errorf("failed to resolve %s: %v", oldest, err)
	}

	logger.Debug("This is the newest commit hash %s", newestHash.String())
	logger.Debug("This is the oldest commit hash %s", oldestHash.String())

	commitsIter, err := repo.Log(&git.LogOptions{From: *newestHash, Order: git.LogOrderCommitterTime})
	if err != nil {
		return commits, fmt.Errorf("failed to git log: %v", err)
	}

	for {
		commit, err := commitsIter.Next()
		if err != nil {
			return commits, fmt.Errorf("failed to get next commit: %v", err)
		}
		if commit.Hash != *oldestHash {
			currentCommit := asArtifactCommit(commit, branchName)
			commits = append(commits, currentCommit)
		} else {
			break
		}
	}

	return commits, nil
}

// asArtifactCommit returns an ArtifactCommit from a git Commit object
func asArtifactCommit(commit *object.Commit, branchName string) *ArtifactCommit {
	var commitParents []string
	for _, h := range commit.ParentHashes {
		commitParents = append(commitParents, h.String())
	}
	return &ArtifactCommit{
		Sha1:      commit.Hash.String(),
		Message:   strings.TrimSpace(commit.Message),
		Author:    commit.Author.String(),
		Timestamp: commit.Author.When.UTC().Unix(),
		Branch:    branchName,
		Parents:   commitParents,
	}
}

// branchName returns the current branch name on a repository,
// or an error if the repo head is not on a branch
func branchName(repo *git.Repository) (string, error) {
	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get the current HEAD of the git repository: %v", err)
	}
	if head.Name().IsBranch() {
		return head.Name().Short(), nil
	}
	return "", nil
}

func artifactCreationDesc() string {
	return `
   Report an artifact creation to a Kosli pipeline. 
   ` + sha256Desc
}
