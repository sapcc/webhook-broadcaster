package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/concourse/concourse/atc"
	"github.com/concourse/concourse/go-concourse/concourse"
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

	teams, err := client.ListTeams()
	if err != nil {
		return fmt.Errorf("Failed to list teams: %s", err)
	}
	pipelinesByID := make(map[int]atc.Pipeline, 50)
	for _, team := range teams {
		client := client.Team(team.Name)
		pipelines, err := client.ListPipelines()
		if err != nil {
			return fmt.Errorf("Failed to list pipelines: %s", err)
		}
		log.Printf("Processing %d pipeline(s) for team %s", len(pipelines), team.Name)

		//update pipeline cache
		for _, pipeline := range pipelines {
			//temporarly memorize pipelines from team to cleanup after the teams loop
			pipelinesByID[pipeline.ID] = pipeline

			config, _, version, found, err := client.PipelineConfig(pipeline.Name)
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
	}
	//delete removed pipelines from cache
	resourceCache.Range(func(key, value interface{}) bool {
		pipelineID := key.(int)
		cachedPipeline := value.(Pipeline)
		if _, found := pipelinesByID[pipelineID]; !found {
			log.Printf("Removing vanished pipeline %s/%s from cache", cachedPipeline.Team, cachedPipeline.Name)
			resourceCache.Delete(pipelineID)
		}
		return true
	})
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
