// Copyright 2016-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package types

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	taskArn          = "arn:aws:ecs:us-east-1:12345678912:task/c024d145-093b-499a-9b14-5baf273f5835"
	instanceArn      = "arn:aws:us-east-1:123456789123:container-instance/4b6d45ea-a4b4-4269-9d04-3af6ddfdc597"
	desiredTaskCount = 5
)

type DeploymentTestSuite struct {
	suite.Suite
	deployment *Deployment
	failures   []*ecs.Failure
	token      string
}

func (suite *DeploymentTestSuite) SetupTest() {
	suite.token = uuid.NewRandom().String()

	var err error
	suite.deployment, err = NewDeployment(taskDefinition, suite.token)
	assert.Nil(suite.T(), err, "Cannot initialize DeploymentTestSuite")

	failedInstance := ecs.Failure{
		Arn: aws.String(instanceArn),
	}
	suite.failures = []*ecs.Failure{&failedInstance}
}

func TestDeploymentTestSuite(t *testing.T) {
	suite.Run(t, new(DeploymentTestSuite))
}

func (suite *DeploymentTestSuite) TestNewDeploymentEmptyTaskDefinition() {
	_, err := NewDeployment("", suite.token)
	assert.Error(suite.T(), err, "Expected an error when task definition is empty")
}

func (suite *DeploymentTestSuite) TestNewDeployment() {
	d, err := NewDeployment(taskDefinition, suite.token)
	assert.Nil(suite.T(), err, "Unexpected error when creating a deployment")
	assert.NotNil(suite.T(), d, "Deployment should not be nil")
	assert.NotEmpty(suite.T(), d.ID, "Deployment ID should not be empty")
	assert.Exactly(suite.T(), DeploymentPending, d.Status, "Deployment status should be pending")
	assert.Exactly(suite.T(), DeploymentHealthy, d.Health, "Deployment should be healthy")
	assert.NotNil(suite.T(), d.StartTime, "Deployment startTime should not be empty")
	assert.Empty(suite.T(), d.EndTime, "Deployment endtime should be empty")
	assert.Exactly(suite.T(), taskDefinition, d.TaskDefinition, "Deployment taskDefintion does not match expected")
}

func (suite *DeploymentTestSuite) TestUpdateDeploymentToInProgressDeploymentCompleted() {
	suite.deployment.Status = DeploymentCompleted

	err := suite.deployment.UpdateDeploymentToInProgress(desiredTaskCount, suite.failures)
	assert.Error(suite.T(), err, "Expected an error when deployment is complete")
}

func (suite *DeploymentTestSuite) TestUpdateDeploymentToInProgressUnhealthy() {
	err := suite.deployment.UpdateDeploymentToInProgress(desiredTaskCount, suite.failures)
	assert.Nil(suite.T(), err, "Unexpected error when setting deployment in progress")
	assert.NotNil(suite.T(), suite.deployment, "Deployment should not be nil")
	assert.NotEmpty(suite.T(), suite.deployment.ID, "Deployment ID should not be empty")
	assert.Exactly(suite.T(), DeploymentInProgress, suite.deployment.Status, "Deployment status should be inprogress")
	assert.Exactly(suite.T(), DeploymentUnhealthy, suite.deployment.Health, "Deployment should be unhealthy")
	assert.Exactly(suite.T(), desiredTaskCount, suite.deployment.DesiredTaskCount, "Deployment desired task count should match expected")
	assert.NotNil(suite.T(), suite.deployment.StartTime, "Deployment startTime should not be empty")
	assert.Empty(suite.T(), suite.deployment.EndTime, "Deployment endtime should be empty")
	assert.Exactly(suite.T(), taskDefinition, suite.deployment.TaskDefinition, "Deployment taskDefintion does not match expected")
	assert.Exactly(suite.T(), suite.failures, suite.deployment.FailedInstances, "Deployment failed instances does not match expected")
}

func (suite *DeploymentTestSuite) TestUpdateDeploymentToInProgressHealthy() {
	err := suite.deployment.UpdateDeploymentToInProgress(desiredTaskCount, []*ecs.Failure{})
	assert.Nil(suite.T(), err, "Unexpected error when setting deployment in progress")
	assert.NotNil(suite.T(), suite.deployment, "Deployment should not be nil")
	assert.NotEmpty(suite.T(), suite.deployment.ID, "Deployment ID should not be empty")
	assert.Exactly(suite.T(), DeploymentInProgress, suite.deployment.Status, "Deployment status should be inprogress")
	assert.Exactly(suite.T(), DeploymentHealthy, suite.deployment.Health, "Deployment should be healthy")
	assert.Exactly(suite.T(), desiredTaskCount, suite.deployment.DesiredTaskCount, "Deployment desired task count should match expected")
	assert.NotNil(suite.T(), suite.deployment.StartTime, "Deployment startTime should not be empty")
	assert.Empty(suite.T(), suite.deployment.EndTime, "Deployment endtime should be empty")
	assert.Exactly(suite.T(), taskDefinition, suite.deployment.TaskDefinition, "Deployment taskDefintion does not match expected")
	assert.Empty(suite.T(), suite.deployment.FailedInstances, "Deployment failed instances does not match expected")
}

func (suite *DeploymentTestSuite) TestUpdateDeploymentToCompletedUnhealthy() {
	suite.deployment.UpdateDeploymentToInProgress(desiredTaskCount, suite.failures)

	err := suite.deployment.UpdateDeploymentToCompleted(suite.failures)
	assert.Nil(suite.T(), err, "Unexpected error when setting deployment to completed")
	assert.NotNil(suite.T(), suite.deployment, "Deployment should not be nil")
	assert.NotEmpty(suite.T(), suite.deployment.ID, "Deployment ID should not be empty")
	assert.Exactly(suite.T(), DeploymentCompleted, suite.deployment.Status, "Deployment status should be completed")
	assert.Exactly(suite.T(), DeploymentUnhealthy, suite.deployment.Health, "Deployment should not be healthy")
	assert.Exactly(suite.T(), desiredTaskCount, suite.deployment.DesiredTaskCount, "Deployment desired task count should match expected")
	assert.NotNil(suite.T(), suite.deployment.StartTime, "Deployment startTime should not be empty")
	assert.NotNil(suite.T(), suite.deployment.EndTime, "Deployment endtime should not be empty")
	assert.Exactly(suite.T(), taskDefinition, suite.deployment.TaskDefinition, "Deployment taskDefintion does not match expected")
	assert.Exactly(suite.T(), suite.failures, suite.deployment.FailedInstances, "Deployment failed instances does not match expected")
}

func (suite *DeploymentTestSuite) TestUpdateDeploymentToCompletedHealthy() {
	suite.deployment.UpdateDeploymentToInProgress(desiredTaskCount, suite.failures)

	err := suite.deployment.UpdateDeploymentToCompleted(nil)
	assert.Nil(suite.T(), err, "Unexpected error when setting deployment to completed")
	assert.NotNil(suite.T(), suite.deployment, "Deployment should not be nil")
	assert.NotEmpty(suite.T(), suite.deployment.ID, "Deployment ID should not be empty")
	assert.Exactly(suite.T(), DeploymentCompleted, suite.deployment.Status, "Deployment status should be completed")
	assert.Exactly(suite.T(), DeploymentHealthy, suite.deployment.Health, "Deployment should be healthy")
	assert.Exactly(suite.T(), desiredTaskCount, suite.deployment.DesiredTaskCount, "Deployment desired task count should match expected")
	assert.NotNil(suite.T(), suite.deployment.StartTime, "Deployment startTime should not be empty")
	assert.NotNil(suite.T(), suite.deployment.EndTime, "Deployment endtime should not be empty")
	assert.Exactly(suite.T(), taskDefinition, suite.deployment.TaskDefinition, "Deployment taskDefintion does not match expected")
	assert.Empty(suite.T(), suite.deployment.FailedInstances, "Deployment failed instances does not match expected")
}
