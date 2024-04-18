package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Workflow struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

type WorkflowRun struct {
	ID         int `json:"id"`
	WorkflowID int `json:"workflow_id"`
}

type WorkflowRunsResponse struct {
	TotalCount   int           `json:"total_count"`
	WorkflowRuns []WorkflowRun `json:"workflow_runs"`
}

func getWorkflows(token, owner, repo, workflowName string, deleteLevel string) []Workflow {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/workflows", owner, repo)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	var data struct {
		Workflows []Workflow `json:"workflows"`
	}
	json.NewDecoder(resp.Body).Decode(&data)

	var filtered []Workflow
	for _, workflow := range data.Workflows {
		if deleteLevel == "repo" {
			if strings.Contains(workflow.State, "disabled") {
				filtered = append(filtered, workflow)
			}

		}
		if deleteLevel == "workflow" {
			if strings.Contains(workflow.Name, workflowName) && strings.Contains(workflow.State, "disabled") {
				filtered = append(filtered, workflow)
			}
		}
	}
	return filtered
}

func fetchRuns(token, owner, repo string, workflowID, page, perPage int) []WorkflowRun {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs?per_page=%d&page=%d", owner, repo, perPage, page)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	var data WorkflowRunsResponse
	json.NewDecoder(resp.Body).Decode(&data)

	var filtered []WorkflowRun
	for _, run := range data.WorkflowRuns {
		if run.WorkflowID == workflowID {
			filtered = append(filtered, run)
		}
	}
	return filtered
}

func getWorkflowRuns(token, owner, repo string, workflowID int) []WorkflowRun {
	page := 1
	perPage := 100
	var filtered []WorkflowRun

	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs?per_page=%d&page=%d", owner, repo, perPage, page)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	var data WorkflowRunsResponse
	json.NewDecoder(resp.Body).Decode(&data)

	totalCount := data.TotalCount
	totalPages := (totalCount / perPage) + 1

	var wg sync.WaitGroup
	var mu sync.Mutex

	for page := 1; page <= totalPages; page++ {
		wg.Add(1)
		go func(page int) {
			defer wg.Done()
			runs := fetchRuns(token, owner, repo, workflowID, page, perPage)
			mu.Lock()
			filtered = append(filtered, runs...)
			mu.Unlock()
		}(page)
	}

	wg.Wait()
	return filtered
}

func deleteSingleRun(token, owner, repo string, runID int) (int, int) {
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/actions/runs/%d", owner, repo, runID)
	req, _ := http.NewRequest("DELETE", url, nil)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	defer resp.Body.Close()

	return resp.StatusCode, runID
}

type DeleteResult struct {
	StatusCode int
	RunID      int
}

func deleteWorkflowRun(token, owner, repo string, runs []WorkflowRun) []DeleteResult {
	var wg sync.WaitGroup
	resultChan := make(chan DeleteResult, len(runs))

	for _, run := range runs {
		wg.Add(1)
		go func(run WorkflowRun) {
			defer wg.Done()
			statusCode, runID := deleteSingleRun(token, owner, repo, run.ID)
			resultChan <- DeleteResult{StatusCode: statusCode, RunID: runID}
		}(run)
	}

	wg.Wait()
	close(resultChan)
	var needDelete []DeleteResult
	for result := range resultChan {

		if result.StatusCode == 204 {
			fmt.Printf("Run ID: %d Deletion Status Code: %d\n", result.RunID, result.StatusCode)
		} else {
			needDelete = append(needDelete, result)
			fmt.Printf("Failed to delete run with ID: %d\n", result.RunID)
		}
	}

	return needDelete
}

var (
	deleteLevel  string
	token        string
	owner        string
	repo         string
	workflowName string
)

func main() {
	flag.StringVar(&deleteLevel, "deleteLevel", "", "Enter the delete level (workflow/repo)")
	flag.StringVar(&token, "token", "", "Enter the personal github access token like (xxx_IygV6o2BhmcZjYHp2AAGtsmmOF0VcV0khxx)")
	flag.StringVar(&owner, "owner", "", "Enter the repo owner name，eg http://www.github.com/owner/repo, owner is the owner")
	flag.StringVar(&repo, "repo", "", "Enter the repo name, eg http://www.github.com/owner/repo, repo is the repo name")
	flag.StringVar(&workflowName, "workflowName", "", "Enter the workflow name (on the left side of the workflow name in the github actions page, eg: CI/CD, can be empty when deleteLevel is repo)")

	flag.Parse()
	if os.Getenv("GITHUB_DELETE_LEVEL") != "" {
		deleteLevel = os.Getenv("GITHUB_DELETE_LEVEL")
	}

	if os.Getenv("GITHUB_TOKEN") != "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if os.Getenv("GITHUB_OWNER") != "" {
		owner = os.Getenv("GITHUB_OWNER")
	}
	if os.Getenv("GITHUB_REPO") != "" {
		repo = os.Getenv("GITHUB_REPO")

	}
	if os.Getenv("GITHUB_WORKFLOW_NAME") != "" {
		workflowName = os.Getenv("GITHUB_WORKFLOW_NAME")
	}

	deleteLevel = strings.TrimSpace(deleteLevel)
	if deleteLevel != "workflow" && deleteLevel != "repo" {
		fmt.Println("Please provide the correct delete level")
		return
	}

	token = strings.TrimSpace(token)
	owner = strings.TrimSpace(owner)
	repo = strings.TrimSpace(repo)

	if deleteLevel == "workflow" {
		workflowName = strings.TrimSpace(workflowName)
		if token == "" || owner == "" || repo == "" || workflowName == "" {
			fmt.Println("Please provide all the required inputs: token, owner, repo, workflowName")
			return
		}
	}

	if deleteLevel == "repo" {
		if token == "" || owner == "" || repo == "" {
			fmt.Println("Please provide all the required inputs: token, owner, repo")
			return
		}
	}

	filteredWorkflows := getWorkflows(token, owner, repo, workflowName, deleteLevel)
	if len(filteredWorkflows) == 0 {
		fmt.Println("No workflows found with the given name, please confirm the workflow exist and it has been disabled")
		return
	}
	for _, workflow := range filteredWorkflows {
		fmt.Printf("%+v\n", workflow)
		workflowID := workflow.ID
		retry := 0
		for {
			if retry > 3 {
				break
			}
			runs := getWorkflowRuns(token, owner, repo, workflowID)
			fmt.Println("numbers need to be deleted:", len(runs))
			needReturn := deleteWorkflowRun(token, owner, repo, runs)
			if len(needReturn) == 0 {
				break
			}
			println("retrying to delete failed runs, sleep 1 minutes")
			time.Sleep(60 * time.Second)
			retry++
		}

	}
	println("All runs deleted successfully")
}
