package containerapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/mcarbonne/minimal-server-monitoring/v2/pkg/utils"
)

var (
	ErrListContainers    = errors.New("failed to list containers")
	ErrInspectContainer  = errors.New("failed to inspect container")
	ErrContainerNotFound = errors.New("container not found")
)

type Client struct {
	http *http.Client
}

func NewClient() (*Client, error) {
	httpc := http.Client{
		Transport: &http.Transport{
			MaxIdleConns:    6,
			IdleConnTimeout: 30 * time.Second,
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", "/var/run/docker.sock")
			},
		},
	}

	return &Client{http: &httpc}, nil
}

func (c *Client) ContainerList(ctx context.Context) ([]Container, error) {
	//nolint:bodyclose // SafeClose instead of Close
	resp, err := c.http.Get("http://localhost/containers/json")
	if err != nil {
		return nil, err
	}
	defer utils.SafeClose(resp.Body)

	if resp.StatusCode != 200 {
		return nil, ErrListContainers
	}

	var result ContainerList

	err = json.NewDecoder(resp.Body).Decode(&result)

	return result, err
}

func (c *Client) ContainerInspect(ctx context.Context, containerId string) (ContainerInspect, error) {
	//nolint:bodyclose // SafeClose instead of Close
	resp, err := c.http.Get("http://localhost/containers/" + containerId + "/json")
	if err != nil {
		return ContainerInspect{}, err
	}
	defer utils.SafeClose(resp.Body)

	if resp.StatusCode == 404 {
		return ContainerInspect{}, fmt.Errorf("%w: '%v'", ErrContainerNotFound, containerId)
	}
	if resp.StatusCode != 200 {
		return ContainerInspect{}, fmt.Errorf("%w: '%v'", ErrInspectContainer, containerId)
	}

	var result ContainerInspect

	err = json.NewDecoder(resp.Body).Decode(&result)

	return ContainerInspect{}, err
}
