package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/docker/docker/api/types"

	"kek/pkg/client"
)

type config struct {
	repo  string
	user  string
	token string
}

func requireFlags() (config, error) {
	repo := flag.String("repo", "", "")
	token := flag.String("token", "", "")
	flag.Parse()
	if *repo == "" || *token == "" {
		return config{}, errors.New("both repo and token are required")
	}
	split := strings.Split(*repo, "/")
	user, repoName := split[0], split[1]
	return config{
		user:  user,
		repo:  repoName,
		token: *token,
	}, nil
}

func main() {
	config, err := requireFlags()
	if err != nil {
		panic(err)
	}
	if err := run(config); err != nil {
		panic(err)
	}

}

func run(args config) error {
	gh := client.NewGithub(args.token)
	valid, err := gh.ValidateToken()
	if err != nil {
		return err
	}
	if !valid {
		return errors.New("provided github token is not valid")
	}
	config, err := gh.GetCIConfig(args.user, args.repo)
	if err != nil {
		return err
	}
	docker, err := client.NewDocker()
	if err != nil {
		return err
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Minute)
	containerID, err := docker.PrepareContainer(ctx, config.Runner)
	if err != nil {
		return err
	}
	repo, err := gh.GetRepoArchive(args.user, args.repo, client.FormatTarball, "master")
	if err != nil {
		return err
	}
	err = docker.CopyToContainer(ctx, containerID, repo.Archive)
	if err != nil {
		return nil
	}
	err, exitCode := executeCommands(ctx, config.Commands, docker, containerID, repo.Name)
	if err != nil {
		if exitCode != 0 {
			_ = docker.ContainerClean(ctx, containerID)
			log.Println(err)
			os.Exit(exitCode)
		}
		return err
	}
	err = docker.ContainerClean(ctx, containerID)
	if err != nil {
		return err
	}
	return nil
}

func executeCommands(ctx context.Context, commands []string, docker *client.Docker, containerID string, path string) (err error, exitCode int) {
	for _, command := range commands {
		conn, execID, err := docker.ContainerExecCommand(ctx, containerID, command, path)
		if err != nil {
			return err, 0
		}
		handleContainerConn(conn)
		inspect, err := docker.ContainerExecInspect(ctx, execID)
		if err != nil {
			return err, 0
		}
		if inspect.ExitCode != 0 {
			return errors.New("unexpected exit"), inspect.ExitCode
		}
	}
	return nil, 0
}

func handleContainerConn(conn types.HijackedResponse) {
	_, err := conn.Reader.WriteTo(os.Stdout)
	if err != nil {
		if err == io.EOF {
			conn.Close()
		}
	}
}
