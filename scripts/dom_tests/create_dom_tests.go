package main

import (
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

const newTestsPath = "/home/aaron/Dev/go/src/github.com/AaronStGeorge/cs-6120/test/df"
const originalTestsPath = "/home/aaron/Dev/misc/bril/examples/test/df"
const turntToml = "command = \"bril2json < {filename} | ../../bin/df {args}\""

func main() {
	err := os.RemoveAll(newTestsPath)
	if err != nil {
		log.Fatalln(err)
	}

	err = os.Mkdir(newTestsPath, 0755)
	if err != nil {
		log.Fatalln(err)
	}

	var paths []string
	err = filepath.WalkDir(originalTestsPath, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() && strings.HasSuffix(path, ".bril") {
			paths = append(paths, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}

	// Write new test file for every command we are interested in
	cmds := []string{"defined", "live", "cprop"}
	for _, cmd := range cmds {
		for _, p := range paths {
			contents, err := ioutil.ReadFile(p)
			if err != nil {
				log.Fatalln(err)
			}

			// Replace "live" the previous command with the one we are interested in. I
			// checked there are only "live"s as command arguments nothing in a program.
			newTestSource := strings.Replace(string(contents), "live", cmd, 1)
			newTestPathNoExtension := newTestsPath + "/" + strings.Split(path.Base(p), ".")[0] + "-" + cmd
			err = os.WriteFile(newTestPathNoExtension+".bril", []byte(newTestSource), 0755)
			if err != nil {
				log.Fatalln(err)
			}

			turnt := exec.Command("turnt", "-vp", "-a", cmd, path.Base(p))
			turnt.Dir = originalTestsPath
			out, err := turnt.Output()
			if err != nil {
				log.Fatalln(err)
			}

			err = os.WriteFile(newTestPathNoExtension+".out", out, 0755)
			if err != nil {
				log.Fatalln(err)
			}
		}
	}

	err = os.WriteFile(newTestsPath+"/turnt.toml", []byte(turntToml), 0755)
	if err != nil {
		log.Fatalln(err)
	}
}
