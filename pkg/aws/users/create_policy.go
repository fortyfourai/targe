package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Permify/targe/internal/ai"
	"github.com/Permify/targe/pkg/aws/models"
	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type CreatePolicy struct {
	controller      *Controller
	lg              *lipgloss.Renderer
	styles          *Styles
	form            *huh.Form
	senderStyle     lipgloss.Style
	err             error
	width           int
	message         *string
	done            *bool
	result          string
	createTerraform *bool
}

func NewCreatePolicy(controller *Controller) CreatePolicy {
	m := CreatePolicy{
		controller: controller,
		width:      maxWidth,
	}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	doneInitialValue := false
	m.done = &doneInitialValue

	messageInitialValue := ""
	m.message = &messageInitialValue

	terraformInitialValue := controller.State.GetTerraform()
	m.createTerraform = &terraformInitialValue

	var groupFields []huh.Field
	groupFields = append(groupFields,
		huh.NewText().
			Key("message").
			Title("Describe Your Policy").
			Value(m.message),
	)
	if !controller.State.GetTerraform() {
		groupFields = append(groupFields,
			huh.NewConfirm().
				Key("terraform").
				Title("Generate Terraform file?").
				Value(m.createTerraform).
				Affirmative("Yes").
				Negative("No"),
		)
	}

	groupFields = append(groupFields,
		huh.NewConfirm().
			Key("done").
			Title("All done?").
			Value(m.done).
			Affirmative("Yes").
			Negative("Refresh"),
	)

	m.form = huh.NewForm(huh.NewGroup(groupFields...)).
		WithWidth(45).
		WithShowHelp(false).
		WithShowErrors(false)

	return m
}

func (m CreatePolicy) Init() tea.Cmd {
	return m.form.Init()
}

func (m CreatePolicy) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, maxWidth) - m.styles.Base.GetHorizontalFrameSize()

	case tea.KeyMsg:

		if msg.String() == "esc" || msg.String() == "ctrl+c" || msg.String() == "q" {
			return m, tea.Quit
		}

		if msg.String() == "c" {
			if m.result != "" {
				err := clipboard.WriteAll(m.result)
				if err != nil {
					m.err = fmt.Errorf("failed to copy snippet: %w", err)
				} else {
					m.err = errors.New("Terraform snippet copied to clipboard!")
				}
			}
		}

		// Check if the "Refresh" or "Done" button was selected
		if msg.String() == "enter" {
			if m.done != nil && *m.done {
				return Switch(m.controller.Next(), 0, 0)
			} else {
				// If no message provided
				if m.message == nil || strings.TrimSpace(*m.message) == "" {
					m.err = errors.New("Please provide a message")
					break
				}

				var resourceArn *string
				if m.controller.State.GetResource() != nil {
					resourceArn = &m.controller.State.GetResource().Arn
				}

				var serviceName *string
				if m.controller.State.GetService() != nil {
					serviceName = &m.controller.State.GetService().Name
				}

				policy, err := ai.GeneratePolicy(
					m.controller.openAiApiKey,
					*m.message,
					serviceName,
					resourceArn,
				)
				if err != nil {
					m.err = err
					break
				}

				// Normal IAM policy JSON
				policyJson, err := json.MarshalIndent(policy, "", "\t")
				if err != nil {
					m.err = err
					break
				}

				if m.createTerraform != nil && *m.createTerraform {
					terraformSnippet := generateTerraformSnippet(policy.Id, string(policyJson), m.controller.State.GetUser().Arn)
					m.result = terraformSnippet
					m.controller.State.SetTerraform(true)
					m.controller.State.SetPolicy(&models.Policy{
						Arn:      "new",
						Name:     policy.Id,
						Document: terraformSnippet,
					})

				} else {
					m.result = string(policyJson)
					m.controller.State.SetPolicy(&models.Policy{
						Arn:      "new",
						Name:     policy.Id,
						Document: string(policyJson),
					})
				}
				m.reinitializeForm()
			}
		}
	}

	var cmds []tea.Cmd

	// Process the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m CreatePolicy) View() string {
	s := m.styles

	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.lg.NewStyle().Margin(1, 0).Render(v)

	var titles []string
	var title string

	if m.controller.State.GetUser() != nil {
		titles = append(titles,
			s.StateHeader.Render("User Name: "+m.controller.State.GetUser().Name),
			s.StateHeader.Render("User ARN: "+m.controller.State.GetUser().Arn),
		)
	}

	if m.controller.State.GetService() != nil && m.controller.State.GetResource() != nil {
		titles = append(titles,
			s.StateHeader.Render("Service Name: "+m.controller.State.GetService().Name),
			s.StateHeader.Render("Resource ARN: "+m.controller.State.GetResource().Arn),
		)
	}

	if len(titles) > 0 {
		title = lipgloss.JoinVertical(lipgloss.Left, titles...)
		title = lipgloss.NewStyle().
			MarginTop(1).
			Render(title)
	}

	// Status (right side)
	var status string
	{
		buildInfo := "(None)"
		if m.result != "" {
			buildInfo = m.result
		}

		const statusWidth = 60
		statusMarginLeft := m.width - statusWidth - lipgloss.Width(form) - s.Status.GetMarginRight()
		status = s.Status.
			Height(lipgloss.Height(form)).
			Width(statusWidth).
			MarginLeft(statusMarginLeft).
			Render(
				s.StatusHeader.Render("Policy") + "\n" + buildInfo,
			)
	}

	errors := m.form.Errors()
	header := lipgloss.JoinVertical(lipgloss.Top,
		m.appBoundaryView("Custom Policy Generator"),
		title,
	)
	if len(errors) > 0 {
		header = m.appErrorBoundaryView(m.errorView())
	}
	body := lipgloss.JoinHorizontal(lipgloss.Top, form, status)

	footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
	if len(errors) > 0 {
		footer = m.appErrorBoundaryView("")
	}

	footerHelp := "Press 'c' to copy snippet to clipboard, ENTER to proceed, ESC to quit."
	footer = lipgloss.JoinVertical(
		lipgloss.Left,
		footer,
		m.appBoundaryView(footerHelp),
	)

	var errMsg string
	if m.err != nil {
		errMsg = m.err.Error()
	}

	var errStyle lipgloss.Style
	if strings.Contains(strings.ToLower(errMsg), "failed") {
		errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // red
	} else {
		errStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
	}

	return s.Base.Render(
		header + "\n" +
			body + "\n\n" +
			footer +
			"\n" +
			errStyle.Render(errMsg),
	)
}

func (m CreatePolicy) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error() + "\n"
	}
	return s
}

func (m CreatePolicy) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (m CreatePolicy) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(red),
	)
}

func (m *CreatePolicy) reinitializeForm() {
	doneInitialValue := false
	m.done = &doneInitialValue

	var groupFields []huh.Field
	groupFields = append(groupFields,
		huh.NewText().
			Key("message").
			Title("Describe Your Policy").
			Value(m.message),
	)

	if m.createTerraform == nil {
		groupFields = append(groupFields,
			huh.NewConfirm().
				Key("terraform").
				Title("Generate Terraform file?").
				Value(m.createTerraform).
				Affirmative("Yes").
				Negative("No"),
		)
	}

	groupFields = append(groupFields,
		huh.NewConfirm().
			Key("done").
			Title("All done?").
			Value(m.done).
			Affirmative("Yes").
			Negative("Refresh"),
	)

	m.form = huh.NewForm(huh.NewGroup(groupFields...)).
		WithWidth(45).
		WithShowHelp(false).
		WithShowErrors(false)
}

func generateTerraformSnippet(policyID, policyDoc, userArn string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "%s" {
  name        = "%s"
  path        = "/"
  description = "Policy generated by Targe"
  policy      = <<EOF
%s
EOF
}

resource "aws_iam_user_policy_attachment" "attachment" {
   user       = "%s"
   policy_arn = aws_iam_policy.%s.arn
}
`, policyID, policyID, policyDoc, userArn, policyID)
}
