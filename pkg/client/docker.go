package client

import (
	"context"
	"io"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/pkg/errors"
)

type Docker struct {
	client *client.Client
}

func NewDocker() (*Docker, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	return &Docker{client: cli}, nil
}

func (d *Docker) PrepareContainer(ctx context.Context, image string) (containerID string, err error) {
	err = d.ImageLoad(ctx, image)
	if err != nil {
		return "", err
	}
	containerID, err = d.ContainerCreate(ctx, image)
	if err != nil {
		return "", err
	}
	err = d.ContainerStart(ctx, containerID)
	if err != nil {
		return containerID, err
	}
	return containerID, nil
}

func (d *Docker) ImageLoad(ctx context.Context, image string) error {
	r, err := d.client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer r.Close()
	_, err = d.client.ImageLoad(ctx, r, false)
	if err != nil {
		return err
	}
	return nil
}

func (d *Docker) ContainerCreate(ctx context.Context, image string) (containerID string, err error) {
	ops := &container.Config{Image: image, AttachStdin: true, OpenStdin: true, AttachStderr: true, AttachStdout: true}
	res, err := d.client.ContainerCreate(ctx, ops, nil, nil, "")
	if err != nil {
		return "", err
	}
	if len(res.Warnings) > 0 {
		return "", errors.Errorf("error: %s with warnings: %s", err, strings.Join(res.Warnings, ","))
	}
	return res.ID, nil
}

func (d *Docker) ContainerStart(ctx context.Context, containerID string) error {
	return d.client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
}

func (d *Docker) ContainerClean(ctx context.Context, containerID string) error {
	err := d.ContainerStop(ctx, containerID)
	if err != nil {
		return err
	}
	err = d.ContainerRemove(ctx, containerID)
	if err != nil {
		return err
	}
	return nil
}

func (d *Docker) ContainerStop(ctx context.Context, containerID string) error {
	return d.client.ContainerStop(ctx, containerID, nil)
}

func (d *Docker) ContainerRemove(ctx context.Context, containerID string) error {
	return d.client.ContainerRemove(ctx, containerID, types.ContainerRemoveOptions{})
}

// ContainerExecCommand executes a command inside a container and returns HijackedResponse
// with Reader to handle output. It is up to the caller to close the connection by calling
// HijackedResponse.Close()
func (d *Docker) ContainerExecCommand(ctx context.Context, containerID, command, workingDir string) (types.HijackedResponse, string, error) {
	execOpts := types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          strings.Split(command, " "),
		WorkingDir:   workingDir,
	}
	exec, err := d.client.ContainerExecCreate(ctx, containerID, execOpts)
	if err != nil {
		return types.HijackedResponse{}, "", err
	}
	execStartCheck := types.ExecStartCheck{Tty: true}
	att, err := d.client.ContainerExecAttach(ctx, exec.ID, execStartCheck)
	if err != nil {
		return att, "", err
	}
	err = d.client.ContainerExecStart(ctx, exec.ID, execStartCheck)
	if err != nil {
		return types.HijackedResponse{}, "", err
	}
	return att, exec.ID, nil
}

func (d *Docker) CopyToContainer(ctx context.Context, containerID string, content io.Reader) error {
	return d.client.CopyToContainer(ctx, containerID, ".", content, types.CopyToContainerOptions{})
}

func (d *Docker) ContainerExecInspect(ctx context.Context, execID string) (types.ContainerExecInspect, error) {
	return d.client.ContainerExecInspect(ctx, execID)
}
