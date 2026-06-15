package main

import (
	"github.com/Tvist1988/dockhop/docker"

	"fmt"
	"io"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/moby/moby/client"
)

// version is set at build time via -ldflags "-X main.version=...".
var version = "dev"

type item string

type styles struct {
	title        lipgloss.Style
	item         lipgloss.Style
	selectedItem lipgloss.Style
	pagination   lipgloss.Style
	help         lipgloss.Style
	quitText     lipgloss.Style
}

type model struct {
	containers list.Model // items on the to-do list
	choice     string
	styles     styles // which to-do list item our cursor is pointing at
	status     string
	cli        *client.Client
}

func main() {

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v", "version":
			fmt.Printf("dockhop %s\n", version)
			return
		}
	}

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func newStyles(darkBG bool) styles {
	var s styles
	s.title = lipgloss.NewStyle().MarginLeft(2)
	s.item = lipgloss.NewStyle().PaddingLeft(4)
	s.selectedItem = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	s.pagination = list.DefaultStyles(darkBG).PaginationStyle.PaddingLeft(4)
	s.help = list.DefaultStyles(darkBG).HelpStyle.PaddingLeft(4).PaddingBottom(1)
	s.quitText = lipgloss.NewStyle().Margin(1, 0, 2, 4)
	return s
}

type itemDelegate struct {
	styles *styles
}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := d.styles.item.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return d.styles.selectedItem.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

func (m *model) updateStyles(isDark bool) {
	m.styles = newStyles(isDark)
	m.containers.Styles.Title = m.styles.title
	m.containers.Styles.PaginationStyle = m.styles.pagination
	m.containers.Styles.HelpStyle = m.styles.help
	m.containers.SetDelegate(itemDelegate{styles: &m.styles})
}

func (i item) FilterValue() string { return string(i) }

func initialModel() model {

	cli, err := client.New(
		client.FromEnv,
	)
	if err != nil {
		fmt.Printf("Docker is unavailable: %v", err)
		os.Exit(1)
	}

	containers, err := docker.FetchContainers(cli)
	if err != nil {
		fmt.Printf("Could not get containers: %v", err)
		os.Exit(1)
	}

	listItems := []list.Item{}
	for _, container := range containers {
		listItems = append(listItems, item(container))
	}

	l := list.New(listItems, itemDelegate{}, 20, 14)
	l.Title = "Выбери контейнер докера"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("r"),            // какая клавиша
				key.WithHelp("r", "refresh"), // что показать: «r  refresh»
			),
		}
	}
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(
				key.WithKeys("r"),                       // какая клавиша
				key.WithHelp("r", "refresh containers"), // что показать: «r  refresh»
			),
		}
	}

	m := model{containers: l, cli: cli}
	m.updateStyles(true)
	return m

}

func (m model) Init() tea.Cmd {
	return tea.RequestBackgroundColor
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.BackgroundColorMsg:
		m.updateStyles(msg.IsDark())
		return m, nil

	case tea.WindowSizeMsg:
		m.containers.SetSize(msg.Width, msg.Height)
		return m, nil

	// Is it a key press?
	case tea.KeyPressMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "enter" key and the space bar toggle the selected state
		// for the item that the cursor is pointing at.
		case "enter":
			i, ok := m.containers.SelectedItem().(item)
			if ok {
				m.choice = string(i)
				return m, docker.ExecShell(string(i))
			}
			return m, tea.Quit
		case "r":
			return m, docker.RefreshContainers(m.cli)
		}

	case docker.ExecFinishedMsg:
		if msg.Err != nil {
			m.status = fmt.Sprintf("shell exited with error %s", msg.Err.Error())
		} else {
			m.status = ""
		}

		m.choice = ""
		return m, nil

	case docker.ContainersRefreshedMsg:
		if msg.Err != nil {
			m.status = fmt.Sprintf("failed to refresh containers: %s", msg.Err.Error())
		} else {
			m.status = ""
		}
		listItems := []list.Item{}
		for _, container := range msg.Items {
			listItems = append(listItems, item(container))
		}
		m.containers.SetItems(listItems)
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	var cmd tea.Cmd
	m.containers, cmd = m.containers.Update(msg)
	return m, cmd
}

func (m model) View() tea.View {
	body := m.containers.View()
	if m.choice != "" {
		return tea.NewView(m.styles.title.Render(fmt.Sprintf("Write exit command for quit from %s", m.choice)))
	}
	if m.status != "" {

		footer := m.styles.help.Render(m.status)
		body = lipgloss.JoinVertical(lipgloss.Left, body, footer)
	}

	// Send the UI for rendering
	return tea.NewView(body)
}
