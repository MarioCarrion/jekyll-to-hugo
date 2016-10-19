// Quick and dirty helper program for:
//
// 1. Removing extra tags in frontmatter
//    Keep only those in the `frontmatterWhitelist` variable
// 2. Fixing permalinks
//    `url: /2016/10/11/something/` ---> `url: /2016/10/11/something.html`
// 3. Replacing "post_url" with "relref"
//    `({% post_url 2016-06-01-june-2016-goals %})` ---> `({{< relref "2015-02-12-book-1-soft-skills.markdown" >}})`
//
// go run main.go "/Users/mario/Repositories/mariocarrion.github.com/content/post"
package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var frontmatterWhitelist = [...]string{
	"date:",
	"description:",
	"title:",
	"url:",
	"image:",
	"image_facebook:"}

var postUrlRegex = regexp.MustCompile("(?P<post_url>{% post_url (?P<name>[a-z0-9_-]+) %})")

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

func clearFrontmatterTags(lines []string) ([]string, int) {
	lastFrontMatterIndex := 0
	frontmatter := false
	frontmatterLines := make([]int, 0)
	urlTokenIndex := -1

	for i, line := range lines {
		if line == "---" {
			if frontmatter == true {
				lastFrontMatterIndex = i
				break
			}

			frontmatter = true
			continue
		}

		blacklisted, urlToken := isFrontMatterBlacklisted(line)
		if blacklisted == true {
			frontmatterLines = append(frontmatterLines, i)
		}
		if urlToken == true && strings.HasSuffix(line, "/") {
			urlTokenIndex = i
		}
	}

	lastFrontMatterIndex -= len(frontmatterLines)

	if urlTokenIndex != -1 {
		lines[urlTokenIndex] = lines[urlTokenIndex][0:len(lines[urlTokenIndex])-1] + ".html"
	}

	for i, frontmatterIndex := range frontmatterLines {
		lines = append(lines[:frontmatterIndex-i], lines[frontmatterIndex+1-i:]...)
	}

	return lines, lastFrontMatterIndex
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

	lines := strings.Split(string(input), "\n")

	// Remove extra tag and fix permalinks
	lines, lastFrontMatterIndex := clearFrontmatterTags(lines)

	// post_url -> relref
	for index, line := range lines[lastFrontMatterIndex:len(lines)] {
		if postUrlRegex.MatchString(line) == true {
			relref := fmt.Sprintf("{{< relref \"${%s}.markdown\" >}}", postUrlRegex.SubexpNames()[2])

			lines[index+lastFrontMatterIndex] = postUrlRegex.ReplaceAllString(line, relref)
		}
	}

	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(path, []byte(output), 0644)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	err := filepath.Walk(strings.Join(os.Args[1:], ""), walkFunc)
	if err != nil {
		log.Fatal(err)
	}
}
