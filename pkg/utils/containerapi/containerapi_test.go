package containerapi_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils/containerapi"
	"gotest.tools/v3/assert"
)

var testContainerList = []string{"container_a", "container_b"}

func setup() {
	// start 2 containers
	for _, name := range testContainerList {
		cmd := exec.Command("docker", "run", "-d", "--name", name, "--init", "alpine:latest", "sleep", "infinity")
		err := cmd.Run()
		if err != nil {
			panic("failed to setup")
		}
	}
}

func teardow() {
	for _, name := range testContainerList {
		cmd := exec.Command("docker", "stop", name)
		err := cmd.Run()
		if err != nil {
			panic("failed to teardown")
		}
		cmd = exec.Command("docker", "rm", name)
		err = cmd.Run()
		if err != nil {
			panic("failed to teardown")
		}
	}
}

func TestMain(m *testing.M) {
	setup()
	exitCode := m.Run()
	teardow()

	// Exit with the proper code
	os.Exit(exitCode)
}

func TestContainerAPIFeatures(t *testing.T) {
	dockerClient, err := containerapi.NewClient()
	assert.Equal(t, err, nil)
	list, err := dockerClient.ContainerList(context.Background())
	assert.Equal(t, err, nil)
	assert.Equal(t, len(list), 2)

	for _, elem := range list {
		assert.Equal(t, len(elem.Names), 1)
		assert.Equal(t, elem.Image, "alpine:latest")
		assert.Equal(t, elem.State, "running")

		inspect, err := dockerClient.ContainerInspect(context.Background(), elem.ID)
		assert.Equal(t, err, nil)
		assert.Equal(t, inspect.RestartCount, 0)
	}

	inspect, err := dockerClient.ContainerInspect(context.Background(), "dummyid")
	assert.Equal(t, inspect, containerapi.ContainerInspect{})
	assert.Equal(t, true, errors.Is(err, containerapi.ErrContainerNotFound))
	assert.Equal(t, "container not found: 'dummyid'", err.Error())
}
