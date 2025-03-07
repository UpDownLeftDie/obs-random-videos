package ui

import (
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Styling
var (
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	itemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA"))

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#7D56F4")).
				Background(lipgloss.Color("#FAFAFA"))

	questionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true)

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Bold(true)
)

// SimpleListModel is a simplified list model for selection
type SimpleListModel struct {
	choices  []string
	paths    []string
	cursor   int
	selected int
	quitting bool
	title    string
}

// Init initializes the model
func (m SimpleListModel) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (m SimpleListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			// Exit the program immediately on Ctrl+C
			fmt.Println("\nOperation cancelled by user")
			os.Exit(0)

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", " ":
			m.selected = m.cursor
			m.quitting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the UI
func (m SimpleListModel) View() string {
	if m.quitting {
		if m.selected >= 0 && m.selected < len(m.choices) {
			return fmt.Sprintf("\n%s %s\n",
				highlightStyle.Render("Selected:"),
				m.choices[m.selected])
		}
		return "Nothing selected\n"
	}

	s := fmt.Sprintf("\n%s\n\n", titleStyle.Render(m.title))

	for i, choice := range m.choices {
		cursor := "  "
		if m.cursor == i {
			cursor = "> "
			choice = selectedItemStyle.Render(choice)
		} else {
			choice = itemStyle.Render(choice)
		}
		s += fmt.Sprintf("%s%s\n", cursor, choice)
	}

	s += "\n" + lipgloss.NewStyle().Faint(true).Render("(Use arrow keys to navigate, Enter to select)")

	return s
}

// RunSelectList displays a selection list and returns the selected item
func RunSelectList(title string, items []string, paths []string, cancelOption string) (string, error) {
	// Add cancel option
	items = append(items, cancelOption)
	paths = append(paths, "")

	model := SimpleListModel{
		choices:  items,
		paths:    paths,
		cursor:   0,
		selected: -1,
		title:    title,
	}

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running selection: %w", err)
	}

	if m, ok := finalModel.(SimpleListModel); ok {
		if m.selected < 0 || m.selected >= len(paths) || m.choices[m.selected] == cancelOption {
			return "", nil
		}
		return paths[m.selected], nil
	}

	return "", fmt.Errorf("invalid model type returned")
}

// YesNoModel represents the model for a yes/no question
type YesNoModel struct {
	question      string
	defaultAnswer bool
	answer        bool
	quitting      bool
	cursor        int
}

// Init initializes the model
func (m YesNoModel) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (m YesNoModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "y", "Y":
			m.answer = true
			m.quitting = true
			return m, tea.Quit
		case "n", "N":
			m.answer = false
			m.quitting = true
			return m, tea.Quit
		case "left", "h":
			m.cursor = 0 // Yes
		case "right", "l":
			m.cursor = 1 // No
		case "enter", " ":
			m.answer = m.cursor == 0
			m.quitting = true
			return m, tea.Quit
		case "ctrl+c", "q", "esc":
			// Exit the program immediately on Ctrl+C
			fmt.Println("\nOperation cancelled by user")
			os.Exit(0)
		}
	}
	return m, nil
}

// View renders the UI
func (m YesNoModel) View() string {
	if m.quitting {
		return ""
	}

	yesStyle := lipgloss.NewStyle()
	noStyle := lipgloss.NewStyle()

	if m.cursor == 0 {
		yesStyle = selectedItemStyle
	} else {
		noStyle = selectedItemStyle
	}

	return fmt.Sprintf("\n%s %s / %s\n\n%s\n",
		questionStyle.Render(m.question),
		yesStyle.Render("Yes"),
		noStyle.Render("No"),
		lipgloss.NewStyle().Faint(true).Render("(Use arrow keys to navigate, Enter to select, or type y/n)"))
}

// RunYesNoQuestion displays a yes/no question and returns the answer
func RunYesNoQuestion(question string, defaultAnswer bool) bool {
	var initialCursor int
	if defaultAnswer {
		initialCursor = 0
	} else {
		initialCursor = 1
	}

	model := YesNoModel{
		question:      question,
		defaultAnswer: defaultAnswer,
		answer:        defaultAnswer,
		cursor:        initialCursor,
	}

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return defaultAnswer
	}

	if m, ok := finalModel.(YesNoModel); ok {
		return m.answer
	}

	return defaultAnswer
}

// ShowQuestion displays a yes/no question to the user
func ShowQuestion(message string, defaultValueFlag bool) bool {
	return RunYesNoQuestion(message, defaultValueFlag)
}

// Item represents a selectable item in a list
type Item struct {
	title       string
	description string
	path        string
}

// Title returns the title of the item
func (i Item) Title() string { return i.title }

// Description returns the description of the item
func (i Item) Description() string { return i.description }

// FilterValue returns the value to filter on
func (i Item) FilterValue() string { return i.title }

// SelectModel represents the model for a selection list
type SelectModel struct {
	list     list.Model
	choice   string
	path     string
	quitting bool
}

// Init initializes the model
func (m SelectModel) Init() tea.Cmd {
	return nil
}

// Update handles user input
func (m SelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q", "esc":
			// Exit the program immediately on Ctrl+C
			fmt.Println("\nOperation cancelled by user")
			os.Exit(0)

		case "enter":
			i, ok := m.list.SelectedItem().(Item)
			if ok {
				m.choice = i.title
				m.path = i.path
			}
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the UI
func (m SelectModel) View() string {
	if m.quitting {
		return fmt.Sprintf("Selected: %s\n", m.choice)
	}
	return "\n" + m.list.View()
}

// SelectFromList displays a selection list and returns the selected item
func SelectFromList(title string, items []Item, cancelOption string) (string, error) {
	// Add cancel option
	items = append(items, Item{title: cancelOption})

	// Setup list
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = selectedItemStyle
	delegate.Styles.NormalTitle = itemStyle

	// Convert items to list items
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = item
	}

	// Set reasonable default dimensions
	width := 80
	height := 20

	// Create the list
	l := list.New(listItems, delegate, width, height)
	l.Title = title
	l.Styles.Title = titleStyle
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(true)

	// Create the model
	model := SelectModel{list: l}

	p := tea.NewProgram(model)
	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running selection: %w", err)
	}

	if m, ok := finalModel.(SelectModel); ok {
		if m.choice == cancelOption || m.choice == "" {
			return "", nil
		}

		return m.path, nil
	}

	return "", fmt.Errorf("invalid model type returned")
}

// CreateTransitionVideoSelector creates a list of items for transition video selection
func CreateTransitionVideoSelector(mediaFiles []string) []Item {
	items := []Item{}

	// Create items from media files
	for _, filePath := range mediaFiles {
		// Get just the file name for display
		var fileName string
		if idx := strings.LastIndex(filePath, "/"); idx >= 0 {
			fileName = filePath[idx+1:]
		} else if idx := strings.LastIndex(filePath, "\\"); idx >= 0 {
			fileName = filePath[idx+1:]
		} else {
			fileName = filePath // Just use the whole path if no separators
		}

		items = append(items, Item{
			title: fileName,
			path:  filePath,
		})
	}

	return items
}

// SelectTransitionVideo prompts the user to select a transition video
func SelectTransitionVideo(mediaFiles []string) (string, error) {
	// Extract file names and paths
	fileNames := make([]string, len(mediaFiles))
	filePaths := make([]string, len(mediaFiles))

	for i, filePath := range mediaFiles {
		// Get just the file name for display
		var fileName string
		if idx := strings.LastIndex(filePath, "/"); idx >= 0 {
			fileName = filePath[idx+1:]
		} else if idx := strings.LastIndex(filePath, "\\"); idx >= 0 {
			fileName = filePath[idx+1:]
		} else {
			fileName = filePath
		}

		fileNames[i] = fileName
		filePaths[i] = filePath
	}

	// Use the simple list selection
	return RunSelectList("Select your transition video", fileNames, filePaths, "CANCEL")
}
