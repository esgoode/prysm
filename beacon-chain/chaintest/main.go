package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"path"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/prysmaticlabs/prysm/beacon-chain/chaintest/backend"
	log "github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

func readTestsFromYaml(yamlDir string) ([]interface{}, error) {
	const chainTestsFolderName = "chain-tests"
	const shuffleTestsFolderName = "shuffle-tests"

	var tests []interface{}

	dirs, err := ioutil.ReadDir(yamlDir)
	if err != nil {
		return nil, fmt.Errorf("could not read yaml tests directory: %v", err)
	}
	for _, dir := range dirs {
		files, err := ioutil.ReadDir(path.Join(yamlDir, dir.Name()))
		if err != nil {
			return nil, fmt.Errorf("could not read yaml tests directory: %v", err)
		}
		for _, file := range files {
			filePath := path.Join(yamlDir, dir.Name(), file.Name())
			data, err := ioutil.ReadFile(filePath) // #nosec
			if err != nil {
				return nil, fmt.Errorf("could not read yaml file: %v", err)
			}
			switch dir.Name() {
			case chainTestsFolderName:
				decoded := &backend.ChainTest{}
				if err := yaml.Unmarshal(data, decoded); err != nil {
					return nil, fmt.Errorf("could not unmarshal YAML file into test struct: %v", err)
				}
				tests = append(tests, decoded)
			case shuffleTestsFolderName:
				decoded := &backend.ShuffleTest{}
				if err := yaml.Unmarshal(data, decoded); err != nil {
					return nil, fmt.Errorf("could not unmarshal YAML file into test struct: %v", err)
				}
				tests = append(tests, decoded)
			}
		}
	}
	return tests, nil
}

func runTests(tests []interface{}, sb *backend.SimulatedBackend) error {
	for _, tt := range tests {
		switch typedTest := tt.(type) {
		case *backend.ChainTest:
			log.Infof("Title: %v", typedTest.Title)
			log.Infof("Summary: %v", typedTest.Summary)
			log.Infof("Test Suite: %v", typedTest.TestSuite)
			for _, testCase := range typedTest.TestCases {
				if err := sb.RunChainTest(testCase); err != nil {
					return fmt.Errorf("chain test failed: %v", err)
				}
			}
		case *backend.ShuffleTest:
			log.Infof("Title: %v", typedTest.Title)
			log.Infof("Summary: %v", typedTest.Summary)
			log.Infof("Test Suite: %v", typedTest.TestSuite)
			log.Infof("Fork: %v", typedTest.Fork)
			log.Infof("Version: %v", typedTest.Version)
			for _, testCase := range typedTest.TestCases {
				if err := sb.RunShuffleTest(testCase); err != nil {
					return fmt.Errorf("chain test failed: %v", err)
				}
			}
		default:
			return fmt.Errorf("receive unknown test type: %T", typedTest)
		}
	}
	return nil
}

func main() {
	var yamlDir = flag.String("tests-dir", "", "path to directory of yaml tests")
	flag.Parse()

	customFormatter := new(prefixed.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	customFormatter.FullTimestamp = true
	log.SetFormatter(customFormatter)

	tests, err := readTestsFromYaml(*yamlDir)
	if err != nil {
		log.Fatalf("Fail to load tests from yaml: %v", err)
	}

	sb, err := backend.NewSimulatedBackend()
	if err != nil {
		log.Fatalf("Could not create backend: %v", err)
	}

	log.Info("----Running Tests----")
	startTime := time.Now()

	err = runTests(tests, sb)
	if err != nil {
		log.Fatalf("Test failed %v", err)
	}

	endTime := time.Now()
	log.Infof("Test Runs Finished In: %v Seconds", endTime.Sub(startTime).Seconds())
}
