package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	binoc "github.com/autamus/binoc/repo"
	builder "github.com/autamus/builder/repo"
	"github.com/autamus/librarian/config"
	"github.com/autamus/librarian/repo"
)

func main() {
	fmt.Println()
	fmt.Print(` _     _ _                    _             
| |   (_) |__  _ __ __ _ _ __(_) __ _ _ __  
| |   | | '_ \| '__/ _' | '__| |/ _' | '_ \ 
| |___| | |_) | | | (_| | |  | | (_| | | | |
|_____|_|_.__/|_|  \__,_|_|  |_|\__,_|_| |_|
`)
	fmt.Printf("Application Version: v%s\n", config.Global.General.Version)
	fmt.Println()

	// Set inital values for Repository
	path := config.Global.Repo.Path
	packagesPath := filepath.Join(path, config.Global.Packages.Path)
	containersPath := filepath.Join(path, config.Global.Containers.Path)
	defaultEnvPath := filepath.Join(path, config.Global.Containers.DefaultEnvPath)
	pagesBranch := config.Global.Repo.PagesBranch
	libraryPath := filepath.Join(path, config.Global.Library.Path)
	templatePath := filepath.Join(path, config.Global.Template.Path)
	currentContainer := config.Global.Containers.Current
	currentSize := config.Global.Containers.Size
	currentVersion := config.Global.Containers.Version
	currentDescription := ""

	// Initialize parser functionality
	binoc, err := binoc.Init(path,
		strings.Split(config.Global.Parsers.Loaded, ","),
		&binoc.RepoGitOptions{
			Name:     config.Global.Git.Name,
			Username: config.Global.Git.Username,
			Email:    config.Global.Git.Email,
			Token:    config.Global.Git.Token,
		},
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("[Generating Docs for " + currentContainer + "]")

	// Get the type of the container from the repository.
	cType, cPath, err := builder.GetContainerType(containersPath, currentContainer)
	if err != nil {
		log.Fatal(err)
	}

	if cType == builder.SpackType {
		// If the container is a spack environment, find the main spec.
		spackEnv, err := builder.ParseSpackEnv(defaultEnvPath, cPath)
		if err != nil {
			log.Fatal(err)
		}
		// Find the path to the main spec package
		specPath, err := builder.FindPackagePath(spackEnv.Spack.Specs[0], packagesPath)
		if err != nil {
			log.Fatal(err)
		}
		// Parse package for main spec
		result, err := binoc.Parse(specPath, false)
		if err != nil {
			log.Fatal(err)
		}
		// Set container description from package.
		currentDescription = result.Package.GetDescription()
	}
	// Override an empty version with latest
	if currentVersion == "" {
		currentVersion = "latest"
	}

	// Pull gh-branch branch to update if possible.
	err = binoc.PullBranch(pagesBranch)
	if err != nil {
		if err != nil && err.Error() != "branch already exists" {
			log.Fatal(err)
		}
	}
	// Switch to gh-pages branch
	err = binoc.SwitchBranch(pagesBranch)
	if err != nil {
		log.Fatal(err)
	}

	// Read in Markdown file for container if exists
	article, err := repo.ParseArticle(libraryPath, currentContainer)
	if err != nil {
		log.Fatal(err)
	}

	// Add metadata to article
	article.AddName(currentContainer)
	article.AddVersion(currentVersion)
	article.AddDescription(currentDescription)
	article.SetSize(currentSize)
	article.SetDate(time.Now())

	err = repo.WriteArticle(libraryPath, templatePath, article)
	if err != nil {
		log.Fatal(err)
	}

	// Commit changes to repository
	err = binoc.Commit(fmt.Sprintf("Update %s to %s at %s",
		currentContainer,
		currentVersion,
		time.Now().String()),
	)
	for err != nil {
		log.Fatal(err)
	}

	// Push changes back to repository
	for isFastForward(binoc.Push()) {
		for isFastForward(binoc.Pull()) {
			err = binoc.Reset()
			if err != nil {
				log.Fatal(err)
			}
		}
		if err != nil {
			log.Fatal(err)
		}
		err = repo.WriteArticle(libraryPath, templatePath, article)
		if err != nil {
			log.Fatal(err)
		}
		// Commit changes to repository
		err = binoc.Commit(fmt.Sprintf("Update %s to %s at %s",
			currentContainer,
			currentVersion,
			time.Now().String()),
		)
		if err != nil {
			log.Fatal(err)
		}
	}
	if err != nil {
		log.Fatal(err)
	}
}

func isFastForward(err error) bool {
	if err != nil {
		if strings.HasPrefix(err.Error(), "non-fast-forward update") ||
			strings.HasPrefix(err.Error(), "command error") {
			return true
		}
		log.Fatal(err)
	}
	return false
}
