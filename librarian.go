package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"time"

	git "github.com/autamus/binoc/repo"
	parser "github.com/autamus/binoc/repo"
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
	// Initialize parser functionality
	parser.Init(strings.Split(config.Global.Parsers.Loaded, ","))

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
		result, err := parser.Parse(specPath)
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
	err = git.PullBranch(path, pagesBranch)
	if err != nil {
		if err != nil && err.Error() != "branch already exists" {
			log.Fatal(err)
		}
	}

	// Switch to gh-pages branch
	err = git.SwitchBranch(path, pagesBranch)
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
	err = git.Commit(path, fmt.Sprintf("Update %s to %s at %s",
		currentContainer,
		currentVersion,
		time.Now().String()),
		config.Global.Git.Name,
		config.Global.Git.Email,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Push changes back to repository
	err = git.Push(path, config.Global.Git.Username, config.Global.Git.Token)
	if err != nil {
		log.Fatal(err)
	}
}
