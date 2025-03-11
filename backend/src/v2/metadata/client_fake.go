// Copyright 2023 The Kubeflow Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package metadata contains types to record/retrieve metadata stored in MLMD
// for individual pipeline steps.

package metadata

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kubeflow/pipelines/backend/src/v2/objectstore"

	"github.com/golang/glog"
	"github.com/kubeflow/pipelines/api/v2alpha1/go/pipelinespec"
	pb "github.com/kubeflow/pipelines/third_party/ml-metadata/go/ml_metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

type FakeClient struct {
	contexts             []*pb.Context
	artifacts            []*pb.Artifact
	artifactIdsToContext map[int64]*pb.Context
}

func intPtr(i int64) *int64 {
	return &i
}

func strPtr(i string) *string {
	return &i
}

func NewFakeClient() *FakeClient {
	fakeClient := &FakeClient{}
	fakeClient.createDummyData()
	return fakeClient
}

func (c *FakeClient) GetPipeline(ctx context.Context, pipelineName, runID, namespace, runResource, pipelineRoot string, storeSessionInfo string) (*Pipeline, error) {
	return nil, nil
}

func (c *FakeClient) GetDAG(ctx context.Context, executionID int64) (*DAG, error) {
	return nil, nil
}

func (c *FakeClient) PublishExecution(ctx context.Context, execution *Execution, outputParameters map[string]*structpb.Value, outputArtifacts []*OutputArtifact, state pb.Execution_State) error {
	return nil
}

func (c *FakeClient) CreateExecution(ctx context.Context, pipeline *Pipeline, config *ExecutionConfig) (*Execution, error) {
	return nil, nil
}
func (c *FakeClient) PrePublishExecution(ctx context.Context, execution *Execution, config *ExecutionConfig) (*Execution, error) {
	return nil, nil
}

func (c *FakeClient) GetExecutions(ctx context.Context, ids []int64) ([]*pb.Execution, error) {
	return nil, nil
}

func (c *FakeClient) GetExecution(ctx context.Context, id int64) (*Execution, error) {
	return nil, nil
}

func (c *FakeClient) GetPipelineFromExecution(ctx context.Context, id int64) (*Pipeline, error) {
	return nil, nil
}

func (c *FakeClient) GetExecutionsInDAG(ctx context.Context, dag *DAG, pipeline *Pipeline, filter bool) (executionsMap map[string]*Execution, err error) {
	return nil, nil
}
func (c *FakeClient) UpdateDAGExecutionsState(ctx context.Context, dag *DAG, pipeline *Pipeline) (err error) {
	return nil
}
func (c *FakeClient) PutDAGExecutionState(ctx context.Context, executionID int64, state pb.Execution_State) (err error) {
	return nil
}
func (c *FakeClient) GetEventsByArtifactIDs(ctx context.Context, artifactIds []int64) ([]*pb.Event, error) {
	return nil, nil
}

func (c *FakeClient) GetArtifactsByID(ctx context.Context, ids []int64) ([]*pb.Artifact, error) {
	var arts []*pb.Artifact
	for _, id := range ids {
		if id < 0 || int(id) >= len(c.artifacts) {
			return nil, fmt.Errorf("Id not found.")
		}
		arts = append(arts, c.artifacts[id])
	}
	return arts, nil
}

func (c *FakeClient) GetArtifactName(ctx context.Context, artifactId int64) (string, error) {
	return "", nil
}

// dummy test data for orderByField/filterQuery/nextPageToken not implemented.
func (c *FakeClient) GetArtifacts(ctx context.Context, maxResultSize int32, orderByAscending bool, orderByField, filterQuery, nextPageToken string) ([]*pb.Artifact, *string, error) {
	if maxResultSize < 0 || int(maxResultSize) > len(c.artifacts) {
		return nil, nil, fmt.Errorf("maxResultSize is out of bounds.")
	}
	artifacts := c.artifacts[:maxResultSize]
	// reverse slice
	if !orderByAscending {
		for i, j := 0, len(artifacts)-1; i < j; i, j = i+1, j-1 {
			artifacts[i], artifacts[j] = artifacts[j], artifacts[i]
		}
	}
	return artifacts, nil, nil
}

func (c *FakeClient) GetContexts(ctx context.Context, maxResultSize int32, orderByAscending bool, orderByField, filterQuery, nextPageToken string) ([]*pb.Context, *string, error) {
	return c.contexts, nil, nil
}

func (c *FakeClient) GetContextByArtifactID(ctx context.Context, id int64) (*pb.Context, error) {
	if id < 0 || int(id) >= len(c.artifactIdsToContext) {
		return nil, fmt.Errorf("Context with id does not exist")
	}
	return c.artifactIdsToContext[id], nil
}

func (c *FakeClient) GetOutputArtifactsByExecutionId(ctx context.Context, executionId int64) (map[string]*OutputArtifact, error) {
	return nil, nil
}

func (c *FakeClient) RecordArtifact(ctx context.Context, outputName, schema string, runtimeArtifact *pipelinespec.RuntimeArtifact, state pb.Artifact_State, bucketConfig *objectstore.Config) (*OutputArtifact, error) {
	return nil, nil
}

func (c *FakeClient) GetOrInsertArtifactType(ctx context.Context, schema string) (typeID int64, err error) {
	return 0, nil
}

func (c *FakeClient) FindMatchedArtifact(ctx context.Context, artifactToMatch *pb.Artifact, pipelineContextId int64) (matchedArtifact *pb.Artifact, err error) {
	return nil, nil
}

// Dummy data will create 2 artifacts that belong to a context with id 1
// The context has pipeline_root, namespace, and bucket_session_info
// custom properties.
func (c *FakeClient) createDummyData() {
	art1 := &pb.Artifact{
		Id:                       intPtr(0),
		Name:                     strPtr("art-0"),
		Type:                     strPtr("1"),
		Uri:                      strPtr("s3://test-bucket/pipeline/some-pipeline-id/task/key0"),
		CreateTimeSinceEpoch:     intPtr(time.Unix(1, 0).UnixMilli()),
		LastUpdateTimeSinceEpoch: intPtr(time.Unix(1, 0).UnixMilli()),
	}
	art2 := &pb.Artifact{
		Id:                       intPtr(1),
		Name:                     strPtr("art-1"),
		Type:                     strPtr("1"),
		Uri:                      strPtr("s3://test-bucket/pipeline/some-pipeline-id/task/key1"),
		CreateTimeSinceEpoch:     intPtr(time.Unix(2, 0).UnixMilli()),
		LastUpdateTimeSinceEpoch: intPtr(time.Unix(2, 0).UnixMilli()),
	}
	ctx1SessInfo := objectstore.S3Params{
		Region:       "test",
		Endpoint:     "test.endpoint",
		DisableSSL:   false,
		SecretName:   "testsecret",
		AccessKeyKey: "testsecretaccesskey",
		SecretKeyKey: "testsecretsecretkey",
	}
	bucketSessionInfo, err := json.Marshal(ctx1SessInfo)
	if err != nil {
		glog.Fatal("failed to marshal fake session info")
	}
	ctx2SessInfo := map[string]string{
		"Region":       "test2",
		"Endpoint":     "test2.endpoint2",
		"DisableSSL":   "false",
		"SecretName":   "testsecret2",
		"AccessKeyKey": "testsecretaccesskey2",
		"SecretKeyKey": "testsecretsecretkey2",
		"FromEnv":      "false",
	}
	sessInfo := &objectstore.SessionInfo{
		Provider: "s3",
		Params:   ctx2SessInfo,
	}
	storeSessionInfo2, err1 := json.Marshal(sessInfo)
	if err1 != nil {
		glog.Fatal("failed to marshal fake session info")
	}

	ctx1 := &pb.Context{
		Id:   intPtr(0),
		Name: strPtr("ctx-0"),
		Type: strPtr("1"),
		CustomProperties: map[string]*pb.Value{
			"pipeline_root":       StringValue("s3://test-bucket"),
			"bucket_session_info": StringValue(string(bucketSessionInfo)),
			"namespace":           StringValue("test-namespace"),
		},
	}
	ctx2 := &pb.Context{
		Id:   intPtr(1),
		Name: strPtr("ctx-1"),
		Type: strPtr("1"),
		CustomProperties: map[string]*pb.Value{
			"pipeline_root":      StringValue("s3://test-bucket"),
			"store_session_info": StringValue(string(storeSessionInfo2)),
			"namespace":          StringValue("test-namespace"),
		},
	}
	c.contexts = []*pb.Context{ctx1, ctx2}
	c.artifacts = []*pb.Artifact{art1, art2}
	c.artifactIdsToContext = map[int64]*pb.Context{
		*art1.Id: ctx1,
		*art2.Id: ctx2,
	}
}
