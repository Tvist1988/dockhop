package docker

import (
	"context"
	"os/exec"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/moby/moby/client"
)

type ExecFinishedMsg struct {
	Err error
}

// в пакете docker
type ContainersRefreshedMsg struct {
	Items []string
	Err   error
}

func FetchContainers(cli *client.Client) ([]string, error) {
	containers, err := cli.ContainerList(context.Background(), client.ContainerListOptions{All: false})
	if err != nil {
		return nil, err
	}
	items := []string{}
	for _, container := range containers.Items {
		name := container.ID // fallback, если имени нет
		if len(container.Names) > 0 {
			name = strings.TrimPrefix(container.Names[0], "/")
		}
		items = append(items, name)
	}

	return items, nil
}

func ExecShell(name string) tea.Cmd {
	c := exec.Command("docker", "exec", "-it", name, "/bin/sh")
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return ExecFinishedMsg{Err: err}
	})
}

func RefreshContainers(cli *client.Client) tea.Cmd {
	return func() tea.Msg {
		items, err := FetchContainers(cli)
		return ContainersRefreshedMsg{Items: items, Err: err}
	}
}
