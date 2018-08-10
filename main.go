package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/concourse/atc"
	"github.com/concourse/go-concourse/concourse"
)

type resource struct {
	team     string
	pipeline string
	name     string
	token    string
}

var (
	listenAddr      string
	concourseURL    string
	authUser        string
	authPassword    string
	refreshInterval time.Duration
)

func init() {
	flag.StringVar(&listenAddr, "listen-addr", ":8080", "Listen address of webhook ingester")
	flag.StringVar(&concourseURL, "concourse-url", "", "External URL of the concourse api")
	flag.StringVar(&authUser, "auth-user", "", "Basic auth concourse username")
	flag.StringVar(&authPassword, "auth-password", "", "Basic auth concourse password")
	flag.DurationVar(&refreshInterval, "refresh-interval", 5*time.Minute, "Resource refresh interval")
}

func main() {
	flag.Parse()

	if concourseURL == "" || authUser == "" || authPassword == "" {
		log.Fatal("Missing one or more of required flags: -concourse-url -auth-user -auth-password")
	}

	bc := basicAuthHttpClient(authUser, authPassword, false, nil)
	basicAuthClient := concourse.NewClient(concourseURL, bc, false)

	go func() {
		for {
			// Todo: reuse token
			if token, err := basicAuthClient.Team("main").AuthToken(); err == nil {
				client := concourse.NewClient(concourseURL, defaultHttpClient(&token, false, nil), false)
				UpdateCache(client)
			} else {
				log.Printf("Failed to authenticate to %s: %s", concourseURL, err)
			}
			time.Sleep(refreshInterval)
		}
	}()

	http.HandleFunc("/github", GithubWebhookHandler)

	log.Printf("Listening for incoming webhooks on %s", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, nil))

}

func GithubWebhookHandler(rw http.ResponseWriter, req *http.Request) {

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

	start := time.Now()
	notifyCount := 0
	err = ScanResourceCache(func(pipeline Pipeline, resource atc.ResourceConfig) (bool, error) {
		if resource.Type != "git" {
			return false, nil
		}
		if uri, ok := resource.Source["uri"].(string); ok {
			if SameGitRepository(uri, pushEvent.Repository.CloneURL) {
				webhookURL := fmt.Sprintf("%s/api/v1/teams/%s/pipelines/%s/resources/%s/check/webhook?webhook_token=%s",
					concourseURL,
					pipeline.Team,
					pipeline.Name,
					resource.Name,
					resource.WebhookToken,
				)
				log.Printf("Notifying resource %s/%s on behalf of %s", pipeline.Name, resource.Name, uri)
				response, err := http.Post(webhookURL, "", nil)
				if err != nil || response.StatusCode >= 400 {
					log.Printf("Failed to notify resource %s/%s. URL: %s, response: %s Error: %v",
						pipeline.Name,
						resource.Name,
						webhookURL,
						response.Status,
						err,
					)
				} else {
					notifyCount++
				}
			}
		}
		return false, nil
	})

	log.Printf("Notified %d resources in %s.", notifyCount, time.Now().Sub(start))

}
