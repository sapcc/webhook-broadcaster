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
	//resourceCache = map[int]Pipeline{}
	resourceCache sync.Map
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
	resourceCache.Range(func(key, _ interface{}) bool {
		pipelineID := key.(int)
		if cachedPipeline, found := pipelinesByID[pipelineID]; !found {
			log.Printf("Removing pipeline %s/%s from cache", cachedPipeline.TeamName, cachedPipeline.Name)
			resourceCache.Delete(pipelineID)
		}
		return true
	})

	//update pipeline cache
	for _, pipeline := range pipelines {
		config, _, version, found, err := client.Team(pipeline.TeamName).PipelineConfig(pipeline.Name)
		if err != nil {
			log.Printf("Failed to get pipeline %s/%s: %s", pipeline.TeamName, pipeline.Name, err)
			continue
		}
		if found {
			cachedPipeline, inCache := resourceCache.Load(pipeline.ID)
			//add or replace cache for pipeline
			if !inCache || cachedPipeline.(Pipeline).Version != version {
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
				resourceCache.Store(pipeline.ID, newCacheObj)
				log.Printf("New version detected for pipeline %s/%s. Found %d resource(s) that have a webhook token.", pipeline.TeamName, pipeline.Name, len(newCacheObj.Resources))
			}
		}
	}
	return nil
}

func ScanResourceCache(walkFn func(pipeline Pipeline, resource atc.ResourceConfig) bool) {
	resourceCache.Range(func(_, val interface{}) bool {
		pipeline := val.(Pipeline)
		for _, resource := range pipeline.Resources {
			if !walkFn(pipeline, resource) {
				return false
			}
		}
		return true
	})
}
