// Quick and dirty helper program for:
// 1. Removing extra tags in frontmatter
//    Keep only those in the `frontmatterWhitelist` variable
// 2. Fixing permalinks
//    `url: /2016/10/11/something/` ---> `url: /2016/10/11/something.html`
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const prefix = "/Users/mario/Repositories/mariocarrion.github.com/content/post"

var frontmatterWhitelist = [...]string{
	"date:",
	"description:",
	"title:",
	"url:",
	"image:",
	"image_facebook:"}

func isFrontMatterBlacklisted(line string) (blacklisted, url bool) {
	for _, value := range frontmatterWhitelist {
		if len(value) > len(line) {
			continue
		}

		if line[:len(value)] == value {
			return false, value == "url:"
		}
	}

	return true, false
}

func walkFunc(path string, info os.FileInfo, err error) error {
	if filepath.Ext(info.Name()) != ".md" && filepath.Ext(info.Name()) != ".markdown" {
		return nil
	}

	fmt.Println("filename is: ", path)

	input, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	frontmatter := false
	frontmatterLines := make([]int, 0)
	urlTokenIndex := -1

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if line == "---" {
			if frontmatter == true {
				frontmatter = false
			} else {
				frontmatter = true
			}
			continue
		}

		if frontmatter == true {
			blacklisted, urlToken := isFrontMatterBlacklisted(line)
			if blacklisted == true {
				frontmatterLines = append(frontmatterLines, i)
			}
			if urlToken == true && strings.HasSuffix(line, "/") {
				urlTokenIndex = i
			}
		}
	}

	if urlTokenIndex != -1 {
		lines[urlTokenIndex] = lines[urlTokenIndex][0:len(lines[urlTokenIndex])-1] + ".html"
	}

	for i, frontmatterIndex := range frontmatterLines {
		lines = append(lines[:frontmatterIndex-i], lines[frontmatterIndex+1-i:]...)
	}

	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(path, []byte(output), 0644)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	err := filepath.Walk(prefix, walkFunc)
	if err != nil {
		log.Fatal(err)
	}
}
