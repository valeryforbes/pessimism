package pipeline

import (
	"context"
	"fmt"
	"testing"

	"github.com/base-org/pessimism/internal/core"
	"github.com/base-org/pessimism/internal/etl/component"
	pipe_registry "github.com/base-org/pessimism/internal/etl/registry/pipe"
	"github.com/stretchr/testify/assert"
)

var (
	nilPID = core.MakePipelineUUID(0, core.NilComponentUUID(), core.NilComponentUUID())

	cID1 = core.MakeComponentUUID(0, 0, 0, 0)
	cID2 = core.MakeComponentUUID(0, 0, 0, 0)
)

// getTestPipeLine ... Returns a test pipeline
func getTestPipeLine(ctx context.Context) Pipeline {

	p1, err := pipe_registry.NewEventParserPipe(ctx, &core.ClientConfig{}, component.WithCUUID(cID1))
	if err != nil {
		panic(err)
	}

	p2, err := pipe_registry.NewEventParserPipe(ctx, &core.ClientConfig{}, component.WithCUUID(cID2))
	if err != nil {
		panic(err)
	}

	comps := []component.Component{
		p1,
		p2,
	}

	pipeLine, err := NewPipeline(&core.PipelineConfig{}, core.NilPipelineUUID(), comps)
	if err != nil {
		panic(err)
	}

	return pipeLine
}

func Test_EtlStore(t *testing.T) {
	var tests = []struct {
		name        string
		function    string
		description string

		constructionLogic func() EtlStore
		testLogic         func(*testing.T, EtlStore)
	}{
		{
			name:        "Successful Add When PID Already Exists",
			function:    "addPipeline",
			description: "",

			constructionLogic: func() EtlStore {
				ctx := context.Background()

				testRegistry := NewEtlStore()
				testPipeLine := getTestPipeLine(ctx)

				testRegistry.AddPipeline(core.NilPipelineUUID(), testPipeLine)
				return testRegistry
			},
			testLogic: func(t *testing.T, store EtlStore) {
				ctx := context.Background()
				testPipeLine := getTestPipeLine(ctx)

				pID2 := core.MakePipelineUUID(
					0,
					core.MakeComponentUUID(0, 0, 0, 1),
					core.MakeComponentUUID(0, 0, 0, 1),
				)

				store.AddPipeline(pID2, testPipeLine)

				for _, comp := range testPipeLine.Components() {
					pIDs, err := store.GetPipelineUUIDs(comp.UUID())

					assert.NoError(t, err)
					assert.Len(t, pIDs, 2)
					assert.Equal(t, pIDs[0].PID, nilPID.PID)
					assert.Equal(t, pIDs[1].PID, pID2.PID)
				}

			},
		},
		{
			name:        "Successful Add When PID Does Not Exists",
			function:    "addPipeline",
			description: "",

			constructionLogic: func() EtlStore {
				pr := NewEtlStore()
				return pr
			},
			testLogic: func(t *testing.T, store EtlStore) {
				ctx := context.Background()
				testPipeLine := getTestPipeLine(ctx)

				pID := core.MakePipelineUUID(
					0,
					core.MakeComponentUUID(0, 0, 0, 1),
					core.MakeComponentUUID(0, 0, 0, 1),
				)

				store.AddPipeline(pID, testPipeLine)

				for _, comp := range testPipeLine.Components() {
					pIDs, err := store.GetPipelineUUIDs(comp.UUID())

					assert.NoError(t, err)
					assert.Len(t, pIDs, 1)
					assert.Equal(t, pIDs[0], pID)
				}

			},
		},
		{
			name:        "Successful Retrieval When CID Is Valid Key",
			function:    "getPipeLineIDs",
			description: "",

			constructionLogic: NewEtlStore,
			testLogic: func(t *testing.T, store EtlStore) {
				cID := core.MakeComponentUUID(0, 0, 0, 0)
				pID := core.MakePipelineUUID(0, cID, cID)

				store.AddComponentLink(cID, pID)

				expectedIDs := []core.PipelineUUID{pID}
				actualIDs, err := store.GetPipelineUUIDs(cID)

				assert.NoError(t, err)
				assert.Equal(t, expectedIDs, actualIDs)

			},
		},
		{
			name:        "Failed Retrieval When CID Is Invalid Key",
			function:    "getPipeLineIDs",
			description: "",

			constructionLogic: NewEtlStore,
			testLogic: func(t *testing.T, store EtlStore) {
				cID := core.MakeComponentUUID(0, 0, 0, 0)

				_, err := store.GetPipelineUUIDs(cID)

				assert.Error(t, err)
			},
		},
		{
			name:        "Failed Retrieval When PID Is Invalid Key",
			function:    "getPipeline",
			description: "",

			constructionLogic: NewEtlStore,
			testLogic: func(t *testing.T, store EtlStore) {
				cID := core.MakeComponentUUID(0, 0, 0, 0)
				pID := core.MakePipelineUUID(0, cID, cID)

				_, err := store.GetPipelineFromPUUID(pID)
				assert.Error(t, err)
				assert.Equal(t, err.Error(), fmt.Sprintf(pIDNotFoundErr, pID))

			},
		}, {
			name:        "Failed Retrieval When Matching UUID Cannot Be Found",
			function:    "getPipeline",
			description: "",

			constructionLogic: func() EtlStore {
				store := NewEtlStore()
				return store
			},
			testLogic: func(t *testing.T, store EtlStore) {
				cID := core.MakeComponentUUID(0, 0, 0, 0)
				pID := core.MakePipelineUUID(0, cID, cID)

				pLine := getTestPipeLine(context.Background())

				store.AddPipeline(pID, pLine)

				pID2 := core.MakePipelineUUID(0, cID, cID)
				_, err := store.GetPipelineFromPUUID(pID2)

				assert.Error(t, err)
				assert.Equal(t, err.Error(), uuidNotFoundErr)

			},
		}, {
			name:        "Successful Retrieval",
			function:    "getPipeline",
			description: "",

			constructionLogic: func() EtlStore {
				store := NewEtlStore()
				return store
			},
			testLogic: func(t *testing.T, store EtlStore) {
				cID := core.MakeComponentUUID(0, 0, 0, 0)
				pID := core.MakePipelineUUID(0, cID, cID)

				expectedPline := getTestPipeLine(context.Background())

				store.AddPipeline(pID, expectedPline)

				actualPline, err := store.GetPipelineFromPUUID(pID)

				assert.NoError(t, err)
				assert.Equal(t, expectedPline, actualPline)
			},
		},
		{
			name:        "Successful Pipeline Fetch",
			function:    "getAllPipelines",
			description: "",

			constructionLogic: func() EtlStore {
				store := NewEtlStore()
				return store
			},
			testLogic: func(t *testing.T, store EtlStore) {
				cID := core.MakeComponentUUID(0, 0, 0, 0)
				pID := core.MakePipelineUUID(0, cID, cID)

				expectedPline := getTestPipeLine(context.Background())

				store.AddPipeline(pID, expectedPline)

				pipelines := store.GetAllPipelines()

				assert.Len(t, pipelines, 1)
				assert.Equal(t, pipelines[0], expectedPline)
			},
		},
	}

	for i, tc := range tests {
		t.Run(fmt.Sprintf("%d-%s-%s", i, tc.function, tc.name), func(t *testing.T) {
			testRouter := tc.constructionLogic()
			tc.testLogic(t, testRouter)
		})

	}
}
