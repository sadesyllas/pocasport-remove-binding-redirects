package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

const regexFormat = `(?ms)(?:` +
	`(\s|\n)+<assemblyBinding[^>]*>(\s|\n)+<dependentAssembly>(\s|\n)+<assemblyIdentity name="%s"[^/]+/>(\s|\n)+<bindingRedirect oldVersion="0.0.0.0-[0-9]+.[0-9]+.[0-9]+.[0-9]+" newVersion="[0-9]+.[0-9]+.[0-9]+.[0-9]+" />(\s|\n)+</dependentAssembly>(\s|\n)+</assemblyBinding>` +
	`|` +
	`(\s|\n)+<dependentAssembly>(\s|\n)+<assemblyIdentity name="%s"[^/]+/>(\s|\n)+<bindingRedirect oldVersion="0.0.0.0-[0-9]+.[0-9]+.[0-9]+.[0-9]+" newVersion="[0-9]+.[0-9]+.[0-9]+.[0-9]+" />(\s|\n)+</dependentAssembly>` +
	`)`

func main() {
	outputSuffix := ""
	if slices.Contains(os.Args, "--debug") {
		outputSuffix = "-debug"
	}

	configPath := path.Join(".", "pocasport-remove-binding-redirects.txt")
	stat, err := os.Stat(configPath)
	if err != nil {
		log.Fatal(err)
	}
	if !stat.Mode().IsRegular() {
		log.Fatalf("%s is not a regular file", configPath)
	}

	fileBytes, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(fileBytes), "\n")

	regexes := make([]*regexp.Regexp, 0)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		regex := fmt.Sprintf(regexFormat, line, line)
		regexes = append(regexes, regexp.MustCompile(regex))
	}

	err = filepath.WalkDir(".", func(filePath string, d os.DirEntry, err error) error {
		if err != nil {
			log.Fatal(err)
		}

		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == "node_modules" || d.Name() == "packages" {
				return fs.SkipDir
			}

			return nil
		}

		found := false
		switch strings.ToLower(d.Name()) {
		case "app.config", "web.config":
			found = true
		}
		if found {
			log.Printf("Processing %s", filePath)

			fileBytes, err := os.ReadFile(filePath)
			if err != nil {
				log.Fatal(err)
			}

			for _, regex := range regexes {
				fileBytes = regex.ReplaceAll(fileBytes, []byte(""))
			}

			log.Printf("Writing %s", filePath+outputSuffix)
			err = os.WriteFile(filePath+outputSuffix, fileBytes, 0644)
			if err != nil {
				log.Fatal(err)
			}
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}
