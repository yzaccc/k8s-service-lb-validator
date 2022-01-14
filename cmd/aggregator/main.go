package main

import (
	"fmt"
	"github.com/k8sbykeshed/k8s-service-validator/tools"
	"github.com/savaki/jq"
	"go.uber.org/zap"
	"strings"
)

/**
- auth
	- gh auth login
- find job ids
	- gh run list --limit 7
- retrieve logs
	- gh run view --job 4800267514 --log | grep "PASS tests"
	- gh run view --job 4800267514 --log | grep "FAIL: tests"
- process
	- in good format as above
 */

type Workflow struct {
	name   string
	status string
	id     string
	lastRunID string
	lastJobID string
	passedTests map[string]bool
	failedTests map[string]bool
}

var wfmap map[string]*Workflow
const cmd = "/usr/local/bin/gh"
const keyPassInLog = "PASS tests"
const keyFailInLog = "Fail: tests"

func main() {
	// gh auth login

	// point to the right repo
	// get github workflows
	wfmap = make(map[string]*Workflow)
	data, err := tools.RunCmd(cmd, "workflow", "list")
	if err != nil {
		zap.L().Fatal(err.Error())
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		ss := strings.Split(line, "\t")
		wf := &Workflow{name: strings.TrimSpace(ss[0]), status: strings.TrimSpace(ss[1]), id: strings.TrimSpace(ss[2])}
		wfmap[wf.id] = wf
	}

	for id, wf := range wfmap {
		wf.passedTests = make(map[string]bool)
		wf.failedTests = make(map[string]bool)
		//fmt.Printf("processing workflow: %v, workflow ID: %v \n", wf.name, id)
		data, err := tools.RunCmd(cmd, "run", "list", "--workflow", id, "--limit", "1")
		if err != nil {
			//zap.L().Fatal(err.Error())
			fmt.Println(err.Error())
		}
		//fmt.Println(string(data))
		ss := strings.Split(string(data), "\t")
		runID := ss[len(ss)-3]
		wf.lastRunID = runID
		//fmt.Printf("last run ID: %v \n", runID)

		// get job id from run id
		data, err = tools.RunCmd(cmd, "api", fmt.Sprintf(
			"https://api.github.com/repos/K8sbykeshed/k8s-service-validator/actions/runs/%v/jobs", runID))

		// parse json response
		op, err := jq.Parse(".jobs.[0].id")
		lastJobID, err := op.Apply(data)
		if err != nil {
			fmt.Printf("error to parse response: %v \n", err)
		}

		//fmt.Printf("job ID: %v \n", string(lastJobID))
		wf.lastJobID = string(lastJobID)

		// retrieve and process job log
		data, err = tools.RunCmd(cmd, "run", "view", "--job", "4812609116", "--log")
		//fmt.Printf("logs: %v \n", string(data))
		logLines := strings.Split(string(data), "\n")
		for _, line := range logLines {
			if index := strings.Index(line, keyPassInLog); index != -1 {
				l := processLine(line, index)
				if l != "" {
					wf.passedTests[l] = true
				}
			} else if index := strings.Index(line, keyFailInLog); index != -1 {
				l := processLine(line, index)
				if l != "" {
					wf.failedTests[l] = true
				}
			}
		}
		//fmt.Println("passed tests:=========")
		//fmt.Println(wf.passedTests)
		//fmt.Println("failed tests =========")
		//fmt.Println(wf.failedTests)
	}

	// report
	// currently only in std, can be extendable to more endpoints
	report := generateReport()
	fmt.Println(report)
}

func processLine(l string, index int) string {
	if len(l) > index+11 {
		l = l[index+11:]
		ll := strings.Split(l, "/")
		if len(ll) == 2 {
			return ll[1]
		}
	}
	return ""
}

func generateReport() string {
	header := "Aggregated testing report from CIs running with various CNIs and proxies:"
	lines := []string{header}
	for _, wf := range wfmap {
		lines = append(lines, "======================")
		lines = append(lines, wf.name)
		lines = append(lines, fmt.Sprintf("* status: %s, action run ID: %s", wf.status, wf.lastRunID))

		if len(wf.passedTests) != 0 {
			lines = append(lines, "PASSED TESTING:")
			for t, _ := range wf.passedTests {
				lines = append(lines, "- " + t)
			}
		} else {
			lines = append(lines, "PASSED TESTING: NONE")
		}

		if len(wf.failedTests) != 0 {
			lines = append(lines, "FAILED TESTING:")
			for t, _ := range wf.failedTests {
				lines = append(lines, "- " + t)
			}
		} else {
			lines = append(lines, "FAILED TESTING: NONE")
		}

	}
	return strings.Join(lines, "\n")
}