package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
)

type StringParser struct {
	referencedValues map[string]bool
}

func newStringParser() *StringParser {
	return &StringParser{
		referencedValues: map[string]bool{},
	}
}

func main() {
	file, err := os.Open("./.env")
	if err != nil {
		panic(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	stringParser := newStringParser()

	data := map[string]string{}

	for scanner.Scan() {
		line := (scanner.Text())

		lineValues := strings.Split(line, "=")
		key := lineValues[0]
		value := lineValues[1]

		_, exists := data[key]
		if exists {
			log.Fatalf("key: %v can not be declared twice", key)
			return
		}

		data[key] = value
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	for key := range data {
		fmt.Println(getValue(key, data, stringParser))
	}

}

func getValue(key string, data map[string]string, stringParser *StringParser) (string, error) {
	val, exists := data[key]
	if !exists {
		return "", errors.New("key does not exist")
	}

	if contains := strings.Contains(val, "${"); !contains {
		return val, nil
	}

	return resolveInterpolation(val, data, stringParser)

}

func resolveInterpolation(value string, data map[string]string, stringParser *StringParser) (string, error) {
	var resolvedValue []byte

	i := 0
	for i < len(value) {
		if string(value[i]) == "$" && i+1 < len(value) && string(value[i+1]) == "{" {
			end := strings.Index(value[i:], "}")
			if end == -1 {
				return "", errors.New("unmatched placeholder")
			}

			nextKey := value[i+2 : i+end]

			if _, exists := stringParser.referencedValues[nextKey]; exists {
				return "", errors.New("circular dependency detected")
			}

			stringParser.referencedValues[nextKey] = true
			nestedString, err := getValue(nextKey, data, stringParser)
			if err != nil {
				return "", err
			}
			//Remove key so that it can be used multiple times in the same string
			delete(stringParser.referencedValues, nextKey)

			resolvedValue = append(resolvedValue, nestedString...)
			i += end + 1
		} else {
			resolvedValue = append(resolvedValue, value[i])
			i++
		}
	}

	return string(resolvedValue), nil
}
