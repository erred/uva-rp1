package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strconv"
)

// pid gets the first pid that matches by executable name
func pid(name string) (int, error) {
	d, err := os.Open("/proc")
	if err != nil {
		return 0, fmt.Errorf("pid %s: %w", name, err)
	}
	for {
		fis, err := d.Readdir(10)
		if err == io.EOF {
			break
		} else if err != nil {
			return 0, fmt.Errorf("pid %s: %w", name, err)
		}
		for _, fi := range fis {
			pid := fi.Name()
			if !fi.IsDir() {
				continue
			}
			p, err := strconv.ParseInt(pid, 10, 0)
			if err != nil {
				continue
			}
			b, err := ioutil.ReadFile("/proc/" + pid + "/stat")
			if err != nil {
				continue
			}
			pname := b[bytes.IndexRune(b, '(')+1 : bytes.IndexRune(b, ')')]
			if string(pname) == name {
				return int(p), nil
			}
		}
	}
	return 0, fmt.Errorf("pid %s: not found", name)
}

// memory gets current rss of pid in bytes
func memory(pid, pagesize int) (int64, error) {
	b, err := ioutil.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return 0, fmt.Errorf("memory %d: %w", pid, err)
	}
	i := bytes.IndexRune(b, ')')
	rss, err := strconv.ParseInt(string(bytes.Fields(b[i+1:])[21]), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("memory %d: %w", pid, err)
	}
	return int64(pagesize) * rss, nil
}
