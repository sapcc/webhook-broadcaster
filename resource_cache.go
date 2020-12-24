package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/concourse/concourse/atc"
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

func UpdateCache(cclient client) error {
	log.Printf("Starting cache update.")

	client, err := cclient.RefreshClientWithToken()
	if err != nil {
		return fmt.Errorf("Failed to create Concourse client")
	}

	teams, err := client.ListTeams()
	if err != nil {
		return fmt.Errorf("Failed to list teams: %s", err)
	}
	pipelinesByID := make(map[int]atc.Pipeline, 50)

	log.Printf("Updating %d teams.", len(teams))

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

			config, version, found, err := client.PipelineConfig(pipeline.Ref())
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

	log.Printf("Ending cache update.")
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
