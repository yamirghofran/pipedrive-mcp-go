package pipedrive

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// GetPipelines fetches all pipelines from Pipedrive.
func (c *Client) GetPipelines(ctx context.Context) (json.RawMessage, error) {
	return c.getRaw(ctx, "pipelines", nil)
}

// GetStages fetches all stages across all pipelines, merging pipeline_name into each stage.
func (c *Client) GetStages(ctx context.Context) ([]json.RawMessage, error) {
	// First fetch all pipelines
	pipelinesData, err := c.getRaw(ctx, "pipelines", nil)
	if err != nil {
		return nil, err
	}

	var pipelines []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if err := json.Unmarshal(pipelinesData, &pipelines); err != nil {
		return nil, fmt.Errorf("parsing pipelines: %w", err)
	}

	// Fetch stages for each pipeline concurrently
	type pipelineStages struct {
		pipelineName string
		stages       []json.RawMessage
		err          error
	}

	results := make([]pipelineStages, len(pipelines))
	var wg sync.WaitGroup

	for i, p := range pipelines {
		wg.Add(1)
		go func(idx int, pipelineID int, pipelineName string) {
			defer wg.Done()
			stagesData, err := c.getRaw(ctx, fmt.Sprintf("pipelines/%d/stages", pipelineID), nil)
			if err != nil {
				results[idx] = pipelineStages{pipelineName: pipelineName, err: err}
				return
			}

			var stages []json.RawMessage
			if err := json.Unmarshal(stagesData, &stages); err != nil {
				results[idx] = pipelineStages{pipelineName: pipelineName, err: err}
				return
			}
			results[idx] = pipelineStages{pipelineName: pipelineName, stages: stages}
		}(i, p.ID, p.Name)
	}
	wg.Wait()

	// Merge all stages with pipeline_name
	var allStages []json.RawMessage
	for _, r := range results {
		if r.err != nil {
			continue // skip failed pipeline stage fetches
		}
		for _, stageData := range r.stages {
			// Add pipeline_name to each stage
			var stage map[string]interface{}
			if err := json.Unmarshal(stageData, &stage); err != nil {
				continue
			}
			stage["pipeline_name"] = r.pipelineName
			enriched, err := json.Marshal(stage)
			if err != nil {
				continue
			}
			allStages = append(allStages, enriched)
		}
	}

	return allStages, nil
}
