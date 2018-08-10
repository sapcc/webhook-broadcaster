package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/concourse/atc"
	"github.com/concourse/go-concourse/concourse"
)

type Pipeline struct {
	ID        int
	Name      string
	Version   string
	Team      string
	Resources []atc.ResourceConfig
}

var (
	mu            sync.Mutex
	resourceCache = map[int]Pipeline{}
)

func UpdateCache(client concourse.Client) error {

	pipelines, err := client.ListPipelines()
	if err != nil {
		return fmt.Errorf("Failed to list pipelines: %s", err)
	}

	pipelinesByID := make(map[int]atc.Pipeline, len(pipelines))
	for _, pipeline := range pipelines {
		pipelinesByID[pipeline.ID] = pipeline
	}

	//delete removed pipelines from cache
	mu.Lock()
	for pipelineID, _ := range resourceCache {
		if cachedPipeline, found := pipelinesByID[pipelineID]; !found {
			log.Printf("Removing pipeline %s from cache", cachedPipeline.Name)
			delete(resourceCache, pipelineID)
		}
	}
	mu.Unlock()

	//update pipeline cache
	for _, pipeline := range pipelines {
		config, _, version, found, err := client.Team(pipeline.TeamName).PipelineConfig(pipeline.Name)
		if err != nil {
			log.Printf("Failed to get pipeline %s: %s", pipeline.Name, err)
			continue
		}
		if found {
			mu.Lock()
			cachedPipeline, inCache := resourceCache[pipeline.ID]
			mu.Unlock()
			//add or replace cache for pipeline
			if !inCache || cachedPipeline.Version != version {
				newCacheObj := Pipeline{
					ID:      pipeline.ID,
					Name:    pipeline.Name,
					Team:    pipeline.TeamName,
					Version: version,
				}
				for _, resource := range config.Resources {
					//Skip resources without webhook tokens
					if resource.WebhookToken == "" {
						continue
					}
					newCacheObj.Resources = append(newCacheObj.Resources, resource)
				}
				mu.Lock()
				resourceCache[pipeline.ID] = newCacheObj
				mu.Unlock()
				log.Printf("Updated cache for pipeline %s, version %s, got %d resources with webhook tokens", pipeline.Name, version, len(newCacheObj.Resources))
			}
		}
	}
	return nil
}

func ScanResourceCache(walkFn func(pipeline Pipeline, resource atc.ResourceConfig) (bool, error)) error {
	mu.Lock()
	defer mu.Unlock()
	for _, pipeline := range resourceCache {
		for _, resource := range pipeline.Resources {
			mu.Unlock()
			stop, err := walkFn(pipeline, resource)
			mu.Lock()
			if stop || err != nil {
				return err
			}
		}
	}
	return nil
}
