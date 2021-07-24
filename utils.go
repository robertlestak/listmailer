package main

import (
	"bufio"
	"os"
	"strings"
)

func StringToAddresses(s string) ([]string, error) {
	addresses := strings.Split(s, " ")
	return addresses, nil
}

func FileToAddresses(f string) ([]string, error) {
	file, err := os.Open(f)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string
	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}
	file.Close()
	return txtlines, nil
}
