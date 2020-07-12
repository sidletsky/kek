package client

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"kek/pkg/ci"
)

const (
	EndpointGraphQL = "https://api.github.com/graphql" // v4
	EndpointREST    = "https://api.github.com"         // v3
)

type ArchiveFormat string

// ArchiveFormats available in github api
const (
	FormatTarball ArchiveFormat = "tarball"
	FormatZipball ArchiveFormat = "zipball"
)

type Github struct {
	token   string
	Client  *http.Client
	GraphQL *GraphQL
}

func (g *Github) Token() string {
	return g.token
}

func (g *Github) SetToken(token string) {
	g.token = "bearer " + token
}

func NewGithub(token string) *Github {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	graphQL := NewGraphQL()
	g := &Github{Client: client, GraphQL: graphQL}
	g.SetToken(token)
	graphQL.Headers = map[string]string{"Authorization": g.Token()}
	return g
}

type ValidTokenPayload struct {
	Data struct {
		Viewer struct {
			IsViewer bool `json:"isViewer"`
		} `json:"viewer"`
	} `json:"data"`
}

func (g *Github) GetViewer() (ValidTokenPayload, error) {
	query := GraphQLQuery{Query: `{ viewer { isViewer } }`}
	res, err := g.GraphQL.Query(EndpointGraphQL, query)
	if err != nil {
		return ValidTokenPayload{}, err
	}
	var response ValidTokenPayload
	if err = json.Unmarshal(res, &response); err != nil {
		return ValidTokenPayload{}, err
	}
	return response, nil
}

func (g *Github) GetCIConfig(user, repo string) (ci.Config, error) {
	content, err := g.GetFileContent(user, repo, ci.ConfigPath)
	if err != nil {
		return ci.Config{}, err
	}
	return parseConfig(content)
}

func parseConfig(config string) (ci.Config, error) {
	var conf ci.Config
	err := yaml.Unmarshal([]byte(config), &conf)
	if err != nil {
		return ci.Config{}, nil
	}
	return conf, err
}

func (g *Github) ValidateToken() (valid bool, err error) {
	viewer, err := g.GetViewer()
	if err != nil {
		return false, err
	}
	return viewer.Data.Viewer.IsViewer, nil
}

type repositoryPayload struct {
	Data struct {
		Repository struct {
			Content struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"repository"`
	} `json:"data"`
}

const getFileContentQuery = `query ($name: String!, $owner: String!, $file: String!) {
  repository(name: $name, owner: $owner) {
    content: object(expression: $file) {
      ... on Blob {
        text
      }
    }
  }
}
`

func (g *Github) GetFileContent(user, repo, file string) (string, error) {
	query := GraphQLQuery{
		Query: getFileContentQuery,
		Variables: map[string]string{
			"name":  repo,
			"owner": user,
			"file":  file,
		},
	}
	res, err := g.GraphQL.Query(EndpointGraphQL, query)
	if err != nil {
		return "", err
	}
	var m repositoryPayload
	err = json.Unmarshal(res, &m)
	if err != nil {
		return "", err
	}
	return m.Data.Repository.Content.Text, nil
}

type RepoArchive struct {
	Archive io.ReadCloser
	Name    string
}

func parseFilename(header http.Header) (string, error) {
	_, param, err := mime.ParseMediaType(header.Get("Content-Disposition"))
	if err != nil {
		return "", err
	}
	filename := param["filename"]
	name := strings.TrimSuffix(filename, ".tar.gz")
	return "/" + name, err
}

// GetRepoArchive returns a git repo archive according to
// https://developer.github.com/v3/repos/contents/#get-archive-link
// It is up to the user to close the reader.
func (g *Github) GetRepoArchive(owner, repo string, format ArchiveFormat, branch string) (*RepoArchive, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/%s/%s", EndpointREST, owner, repo, format, branch)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", g.Token())
	res, err := g.Client.Do(req)
	if err != nil {
		return nil, err
	}
	name, err := parseFilename(res.Header)
	if err != nil {
		return nil, err
	}
	r := &RepoArchive{
		Archive: res.Body,
		Name:    name,
	}
	return r, nil
}
