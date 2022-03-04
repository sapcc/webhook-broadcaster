package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/concourse/concourse/atc"
)

type GithubWebhookHandler struct {
	queue *RequestWorkqueue
}

func (gh *GithubWebhookHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {

	var pushEvent struct {
		Repository struct {
			FullName string `json:"full_name"`
			CloneURL string `json:"clone_url"`
			GitURL   string `json:"git_url"`
		} `json:"repository"`
		Commits []struct {
			ID            string   `json:"id"`
			Message       string   `json:"message"`
			AddedFiles    []string `json:"added"`
			RemovedFiles  []string `json:"removed"`
			ModifiedFiles []string `json:"modified"`
		} `json:"commits"`
	}
	if req.Body == nil {
		rw.WriteHeader(400)
		log.Printf("Empty body")
		return
	}
	err := json.NewDecoder(req.Body).Decode(&pushEvent)
	if err != nil {
		rw.WriteHeader(400)
		log.Printf("Failed to parse request body: %s", err)
		return
	}

	log.Printf("Received webhhook for %s", pushEvent.Repository.CloneURL)

	//collect list of changed files
	filesChanged := []string{}
	for _, commit := range pushEvent.Commits {
		filesChanged = append(filesChanged, commit.AddedFiles...)
		filesChanged = append(filesChanged, commit.RemovedFiles...)
		filesChanged = append(filesChanged, commit.ModifiedFiles...)
	}

	ScanResourceCache(func(pipeline Pipeline, resource atc.ResourceConfig) bool {
		if resource.Type != "git" && resource.Type != "pull-request" && resource.Type != "git-proxy" {
			return true
		}
		if uri, ok := resource.Source["uri"].(string); ok {
			if SameGitRepository(uri, pushEvent.Repository.CloneURL) {
				if paths, ok := resource.Source["paths"].([]string); ok && len(paths) > 1 && !matchFiles(paths, filesChanged) {
					log.Printf("Skipping resource %s due to path filter", resource.Name)
					return true
				}
				webhookURL := fmt.Sprintf("%s/api/v1/teams/%s/pipelines/%s/resources/%s/check/webhook?webhook_token=%s",
					concourseURL,
					pipeline.Team,
					pipeline.Name,
					resource.Name,
					resource.WebhookToken,
				)
				gh.queue.Add(webhookURL)
			}
		}
		return true
	})

}

func matchFiles(patterns []string, files []string) bool {
	for _, file := range files {
		for _, pattern := range patterns {
			// direct match
			if file == pattern {
				return true
			}
			// directory match
			if strings.HasSuffix(pattern, "/") && strings.HasPrefix(file, pattern) {
				return true
			}
			// directory without trainling / match
			if strings.HasPrefix(file, pattern+"/") {
				return true
			}
			//last resort glob match
			if ok, _ := filepath.Match(pattern, file); ok {
				return true
			}
		}
	}
	return false
}
