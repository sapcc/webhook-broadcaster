package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

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
		}
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

	ScanResourceCache(func(pipeline Pipeline, resource atc.ResourceConfig) bool {
		if resource.Type != "git" && resource.Type != "pull-request" {
			return true
		}

		uri, ok := resource.Source["uri"].(string)
		if !ok {
			var repository string
			if repository, ok = resource.Source["repository"].(string); ok {
				uri = "https://github.com/" + repository + ".git"
			}
		}

		if ok {
			log.Printf("uri = %v", uri)
			if SameGitRepository(uri, pushEvent.Repository.CloneURL) {
				webhookURL := fmt.Sprintf("%s/api/v1/teams/%s/pipelines/%s/resources/%s/check/webhook?webhook_token=%s",
					concourseURL,
					pipeline.Team,
					pipeline.Name,
					resource.Name,
					resource.WebhookToken,
				)
				gh.queue.Add(webhookURL)
			}
		} else {
			log.Printf("no uri or repository")
		}

		return true
	})

}
