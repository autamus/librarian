package repo

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v2"
)

type Article struct {
	Layout       string    `yaml:"layout"`
	Name         string    `yaml:"name"`
	GitHub       string    `yaml:"github"`
	Versions     []string  `yaml:"versions"`
	Date         time.Time `yaml:"updated_at"`
	Size         string    `yaml:"size"`
	Description  string    `yaml:"description"`
	ContainerURL string    `yaml:"container_url"`
}

// AddName sets the name of an article
func (a *Article) AddName(name string) {
	a.Name = name
	a.Layout = "container"
	a.GitHub = fmt.Sprintf("https://github.com/autamus/registry/blob/main/containers/%s/%s/spack.yaml",
		string(toHyphenCase(name)[0]),
		toHyphenCase(name),
	)
	a.ContainerURL = fmt.Sprintf("https://github.com/orgs/autamus/packages/container/package/%s",
		toHyphenCase(name))
}

// AddVersion adds a version to the article metadata
func (a *Article) AddVersion(new string) {
	found := false
	for _, v := range a.Versions {
		if v == new {
			found = true
		}
	}
	if !found {
		a.Versions = append(a.Versions, new)
	}
}

// AddDescription adds a description to a an article
func (a *Article) AddDescription(new string) {
	if a.Description == "" {
		a.Description = new
	}
}

// SetSize adds/updates the size of a container to a an article
func (a *Article) SetSize(new string) {
	a.Size = new
}

// SetDate sets the published date for an article
func (a *Article) SetDate(new time.Time) {
	a.Date = new
}

func ParseArticle(libraryPath, containerName string) (result Article, err error) {
	err = filepath.Walk(libraryPath, func(path string, info os.FileInfo, err error) error {
		if path == containerName+".md" {
			// Read in file
			fileRaw, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			yamlRaw := strings.Split(string(fileRaw), "---")[1]
			err = yaml.Unmarshal([]byte(yamlRaw), result)
			if err != nil {
				return err
			}
		}
		return nil
	})
	return result, err
}

func WriteArticle(libraryPath, templatePath string, article Article) (err error) {
	// Write YAML Data
	out, err := yaml.Marshal(article)
	if err != nil {
		return err
	}
	// Read in template
	templateContent, err := ioutil.ReadFile(templatePath)
	if err != nil {
		return err
	}

	// Construct Article from Template
	t, err := template.New("article").Funcs(template.FuncMap{
		"toHypen": toHyphenCase,
		"timeString": func(input time.Time) string {
			return input.String()
		},
	}).Parse(string(templateContent))
	if err != nil {
		return err
	}

	var buff bytes.Buffer
	err = t.Execute(&buff, article)
	if err != nil {
		return err
	}

	// Write --- YAML --- Article to file
	f, err := os.Create(filepath.Join(libraryPath, article.Name+".md"))
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(fmt.Sprintf("---\n%s\n---\n%s", string(out), buff.String()))
	return err
}

// toHypenCase converts a string to a hyphenated version appropriate
// for the commandline.
func toHyphenCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(str, "${1}-${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}-${2}")
	return strings.ToLower(snake)
}
