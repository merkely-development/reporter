package server

import (
	"fmt"
	"os"
	"path"

	"github.com/merkely-development/reporter/internal/digest"
)

// ServerData represents the harvested server artifacts data
type ServerData struct {
	Digests           map[string]string `json:"digests"`
	CreationTimestamp int64             `json:"creationTimestamp"`
}

// CreateServerArtifactsData creates a list of ServerData for server artifacts at given paths
func CreateServerArtifactsData(paths []string) ([]*ServerData, error) {
	result := []*ServerData{}
	for _, p := range paths {
		digests := make(map[string]string)
		sha256, err := digest.DirSha256(p)
		if err != nil {
			return []*ServerData{}, fmt.Errorf("Failed to get a digest of path %s with error: %v", p, err)
		}
		artifactName := path.Base(p)
		digests[artifactName] = sha256
		ts, err := getPathLastModifiedTimestamp(p)
		if err != nil {
			return []*ServerData{}, fmt.Errorf("Failed to get last modified timestamp of path %s with error: %v", p, err)
		}
		result = append(result, &ServerData{Digests: digests, CreationTimestamp: ts})
	}
	return result, nil
}

// getPathLastModifiedTimestamp returns the last modified timestamp of a directory
func getPathLastModifiedTimestamp(path string) (int64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fileInfo.ModTime().Unix(), nil
}
