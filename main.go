package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func replaceLongName(line string) string {
	sa := strings.Split(line, " ")
	for i := range sa {
		if strings.HasPrefix(sa[i], "name=") && len(sa[i]) > 100+7 {
			if strings.Contains(sa[i], `src/android/java/src/com/googlecode/eyesfree`) {
				sa[i] = strings.Replace(sa[i], `src/android/java/src/com/googlecode/eyesfree`, ``, -1)
			} else {
				panic(fmt.Errorf("ERROR: Long repository name: %d: %s", len(sa[i]), sa[i]))
			}
		}
	}
	return strings.Join(sa, " ")
}

func replaceFetchName(line string) string {
	sa := strings.Split(line, " ")
	for i := range sa {
		if strings.HasPrefix(sa[i], "fetch=") {
			sa[i] = `fetch="ssh://git@github.com:22/LinfinyJapan/"`
		}
	}
	return strings.Join(sa, " ")
}

func replaceRepoName(line string) string {
	sa := strings.Split(line, " ")
	for i := range sa {
		if strings.HasPrefix(sa[i], "name=") {
			sa[i] = strings.Replace(sa[i], "/", "-", -1)
		}
	}
	return strings.Join(sa, " ")
}

func tmpName(dirpath string, filename string, suffix string) string {
	return filepath.Join(dirpath, filename+suffix)
}

func renameManifest(dirpath string, finfo os.FileInfo) error {
	src, err := os.Open(filepath.Join(dirpath, finfo.Name()))
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(tmpName(dirpath, finfo.Name(), ".xx"), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer dst.Close()

	scanner := bufio.NewScanner(src)
	for scanner.Scan() {
		dst.WriteString(replaceFetchName(replaceRepoName(replaceLongName(scanner.Text()))))
		dst.WriteString("\n")
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if err = os.Rename(filepath.Join(dirpath, finfo.Name()), tmpName(dirpath, finfo.Name(), ".org")); err != nil {
		return err
	}

	err = os.Rename(tmpName(dirpath, finfo.Name(), ".xx"), filepath.Join(dirpath, finfo.Name()))
	if err == nil {
		err = os.Remove(tmpName(dirpath, finfo.Name(), ".org"))
	}

	return err
}

func renameManifests(dirPath string) error {
	files, err := ioutil.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			renameManifests(filepath.Join(dirPath, file.Name()))
			continue
		}
		if strings.HasSuffix(file.Name(), ".xml") {
			err := renameManifest(dirPath, file)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func main() {
	if len(os.Args) == 1 {
		log.Fatal(errors.New("Missing argument: manifest_file"))
	}

	dirPath := os.Args[1]
	err := renameManifests(dirPath)
	if err != nil {
		panic(err)
	}
}
