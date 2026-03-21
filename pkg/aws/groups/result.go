package groups

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/Permify/targe/pkg/aws/models"
)

type Result struct {
	controller *Controller
	lg         *lipgloss.Renderer
	styles     *Styles
	form       *huh.Form
	width      int
	value      *bool
	error      error
}

func NewResult(controller *Controller) Result {
	// Initialize the Result with default values
	result := Result{
		width:      maxWidth,
		lg:         lipgloss.DefaultRenderer(),
		controller: controller,
	}

	// Initialize styles
	result.styles = NewStyles(result.lg)

	// Initialize value pointer
	initialValue := false
	result.value = &initialValue

	// Configure the form
	result.form = createForm(result.value)

	return result
}

func (m Result) Init() tea.Cmd {
	return m.form.Init()
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func (m Result) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, 80) - m.styles.Base.GetHorizontalFrameSize()
	case tea.KeyMsg:
		if msg.String() == "esc" || msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

		// Handle "Enter" for exit confirmation
		if m.form.State == huh.StateCompleted && msg.String() == "enter" {
			return m, tea.Quit
		}
	}

	var cmds []tea.Cmd

	// Process the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	// Handle form completion
	if m.form.State == huh.StateCompleted {
		if *m.value {
			if err := m.controller.Done(); err != nil {
				// Handle error without quitting
				m.error = err
				return m, nil // Return updated model without quitting
			}
		} else {
			cmds = append(cmds, tea.Quit)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m Result) View() string {
	if m.form.State == huh.StateCompleted && m.error == nil || m.controller.State.terraform {
		// Success Message with Exit Footer
		successMessage := fmt.Sprintf(
			"\n%s\n\n%s\n",
			lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("10")).
				Render("✔ Operation executed successfully!"),
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("7")).
				Italic(true).
				Render("The requested AWS IAM operation has been completed."),
		)

		exitFooter := lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Italic(true).
			Render("Press Enter to exit.")

		return successMessage + "\n" + exitFooter
	}

	// When not in completed state, display other UI elements
	rows := m.collectOverviewRows()
	t := m.createTable(rows)
	formView := m.lg.NewStyle().Margin(1, 0).Render(strings.TrimSuffix(m.form.View(), "\n\n"))
	header := m.renderHeader()
	footer := m.renderFooter()

	body := lipgloss.JoinVertical(lipgloss.Top, t.Render(), formView)

	// Add error message if present
	if m.error != nil {
		errorView := fmt.Sprintf(
			"\n%s\n\n%s\n",
			lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("9")).
				Render("✖ An error occurred"),
			lipgloss.NewStyle().
				Foreground(lipgloss.Color("1")).
				Italic(true).
				Render(m.error.Error()),
		)
		body = lipgloss.JoinVertical(lipgloss.Top, body, errorView)
	}

	return m.styles.Base.Render(header + "\n" + body + "\n\n" + footer)
}

func (m Result) collectOverviewRows() [][]string {
	var rows [][]string
	state := m.controller.State

	if state.group != nil {
		rows = append(rows, []string{"Group", state.group.Name, state.group.Arn})
	}
	if state.operation != nil {
		rows = append(rows, []string{"Operation", state.operation.Name, state.operation.Desc})
	}
	if state.group != nil {
		rows = append(rows, []string{"Group", state.group.Name, state.group.Arn})
	}
	if state.service != nil {
		rows = append(rows, []string{"Service", state.service.Name, state.service.Desc})
	}
	if state.resource != nil {
		rows = append(rows, []string{"Resource", state.resource.Name, state.resource.Arn})
	}
	if state.policy != nil {
		rows = append(rows, m.formatPolicyRow(state.policy))
	}

	return rows
}

func (m Result) formatPolicyRow(policy *models.Policy) []string {
	if len(policy.Document) > 0 {
		return []string{"Policy", policy.Name, "new"}
	}
	return []string{"Policy", policy.Name, policy.Arn}
}

func (m Result) createTable(rows [][]string) *table.Table {
	return table.New().
		Border(lipgloss.HiddenBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("99"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if col == 0 {
				return m.styles.Base.Foreground(lipgloss.Color("205")).Bold(true)
			}
			return m.styles.Base
		}).
		Rows(rows...)
}

func (m Result) renderHeader() string {
	errors := m.form.Errors()
	if len(errors) > 0 {
		return m.appErrorBoundaryView(m.errorView())
	}
	return m.appBoundaryView("Overview")
}

func (m Result) renderFooter() string {
	errors := m.form.Errors()
	if len(errors) > 0 {
		return m.appErrorBoundaryView("")
	}
	return m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
}

func (m Result) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error() + "\n"
	}
	return s
}

func (m Result) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (m Result) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func createForm(value *bool) *huh.Form {
	confirm := huh.NewConfirm().
		Key("done").
		Title("All done?").
		Validate(func(v bool) error {
			return nil
		}).
		Affirmative("Yes").
		Negative("No").
		Value(value)

	return huh.NewForm(
		huh.NewGroup(confirm),
	).
		WithWidth(45).
		WithShowHelp(false).
		WithShowErrors(false)
}
