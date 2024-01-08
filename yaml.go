package reverseproxy

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"

	yaml "gopkg.in/yaml.v3"
)

func parseYaml(yamlFile string, s *server) (err error) {
	data, err := os.ReadFile(yamlFile)
	if err != nil {
		// fmt.Printf("reader proxy file fail, err = %s\n", err.Error())
		return
	}

	regex := regexp.MustCompile("version: (v[0-9]+)")
	r := regex.Find(data)
	if r != nil {
		version := bytes.TrimPrefix(r, []byte("version: "))
		data = bytes.ReplaceAll(data, []byte("{version}"), version)
	}

	if err = yaml.Unmarshal(data, s); err != nil {
		// fmt.Printf("unmarshal fail, err = %s\n", err.Error())
		return
	}
	// fmt.Printf("s = %v\n", s)
	return
}

func walkYaml(dir string) (files []string, err error) {
	if _, err = os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return
		}
	}
	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})
	return
}
