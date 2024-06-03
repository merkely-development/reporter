package sonar

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type SonarConfig struct {
	OrgKey     string
	ProjectKey string
	APIToken   string
}

type SonarResults struct {
	Component Component `json:"component"`
}

type Component struct {
	Id        string     `json:"id"`
	Key       string     `json:"key"`
	Name      string     `json:"name"`
	Qualifier string     `json:"qualifier"`
	Measures  []Measures `json:"measures"`
}

type Measures struct {
	Metric string `json:"metric"`
	Value  string `json:"value"`
}

func NewSonarConfig(orgKey, projectKey, apiToken string) *SonarConfig {
	return &SonarConfig{
		OrgKey:     orgKey,
		ProjectKey: projectKey,
		APIToken:   apiToken,
	}
}

func (sc *SonarConfig) GetSonarResults() (*SonarResults, error) {
	httpClient := &http.Client{}
	var url string
	var token string

	if sc.OrgKey != "" && sc.ProjectKey != "" && sc.APIToken != "" {
		url = fmt.Sprintf("https://sonarcloud.io/api/measures/component?metricKeys=alert_status%%2Cquality_gate_details%%2Cbugs%%2Csecurity_issues%%2Ccode_smells%%2Ccomplexity%%2Cmaintainability_issues%%2Creliability_issues%%2Ccoverage&component=%s_%s", sc.OrgKey, sc.ProjectKey)
	} else {
		return nil, fmt.Errorf("OrgKey, Project Key and API token must be given to retrieve data from SonarCloud/SonarQube")
	}

	//response, err := httpClient.Get(url)
	request, err := http.NewRequest("GET", url, nil)
	request.Header.Add("Authorization", token)

	response, err := httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	sonarResult := &SonarResults{}
	json.NewDecoder(response.Body).Decode(sonarResult)

	return sonarResult, nil
}