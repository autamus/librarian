package config

import (
	"os"
	"reflect"
	"strings"
)

// Config defines the configuration struct for importing settings from ENV Variables
type Config struct {
	General    general
	Git        git
	Repo       repo
	Parsers    parsers
	Packages   packages
	Containers containers
	Template   template
	Library    library
}

type general struct {
	Version string
}

type repo struct {
	Path        string
	PagesBranch string
}

type git struct {
	Name     string
	Username string
	Email    string
	Token    string
}

type packages struct {
	Path string
}

type containers struct {
	Path           string
	Current        string
	Version        string
	DefaultEnvPath string
	Size           string
}

type parsers struct {
	Loaded string
}

type template struct {
	Path string
}

type library struct {
	Path string
}

var (
	// Global is the configuration struct for the application.
	Global Config
)

func init() {
	defaultConfig()
	parseConfigEnv()
}

func defaultConfig() {
	Global.General.Version = "0.0.4"
	Global.Parsers.Loaded = "spack,shpc"
	Global.Containers.Path = "containers/"
	Global.Containers.DefaultEnvPath = "default.yaml"
	Global.Template.Path = "default.md"
	Global.Packages.Path = "spack/"
	Global.Repo.Path = "."
}

func parseConfigEnv() {
	numSubStructs := reflect.ValueOf(&Global).Elem().NumField()
	for i := 0; i < numSubStructs; i++ {
		iter := reflect.ValueOf(&Global).Elem().Field(i)
		subStruct := strings.ToUpper(iter.Type().Name())

		structType := iter.Type()
		for j := 0; j < iter.NumField(); j++ {
			fieldVal := iter.Field(j).String()
			if fieldVal != "Version" {
				fieldName := structType.Field(j).Name
				for _, prefix := range []string{"LIB", "INPUT"} {
					evName := prefix + "_" + subStruct + "_" + strings.ToUpper(fieldName)
					evVal, evExists := os.LookupEnv(evName)
					if evExists && evVal != fieldVal {
						iter.FieldByName(fieldName).SetString(evVal)
					}
				}
			}
		}
	}
}
