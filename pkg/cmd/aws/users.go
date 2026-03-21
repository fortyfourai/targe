package aws

import (
	"context"
	"fmt"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"

	"github.com/spf13/viper"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"github.com/Permify/targe/pkg/aws/models"

	internalaws "github.com/Permify/targe/internal/aws"
	"github.com/Permify/targe/internal/config"
	pkgusers "github.com/Permify/targe/pkg/aws/users"
	"github.com/Permify/targe/pkg/cmd/common"
)

type Users struct {
	model tea.Model
}

func (m Users) Init() tea.Cmd {
	return m.model.Init() // rest methods are just wrappers for the model's methods
}

func (m Users) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.model.Update(msg)
}

func (m Users) View() string {
	return m.model.View()
}

// NewUsersCommand -
func NewUsersCommand(cfg *config.Config) *cobra.Command {
	command := &cobra.Command{
		Use:   "users",
		Short: "",
		RunE:  users(cfg),
	}

	f := command.Flags()

	f.String("user", "", "user")
	f.String("operation", "", "operation")
	f.String("group", "", "group")
	f.String("policy", "", "policy")
	f.String("resource", "", "resource")
	f.String("service", "", "service")
	f.String("policy-option", "", "policy option")
	f.String("terraform", "", "terraform")

	// SilenceUsage is set to true to suppress usage when an error occurs
	command.SilenceUsage = true

	command.PreRun = func(cmd *cobra.Command, args []string) {
		RegisterUsersFlags(f)
	}

	return command
}

func users(cfg *config.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// Replace "requirements" with the actual path to your folder
		requirementsPath := "requirements"

		// Check if the requirements folder exists
		if !folderExists(requirementsPath) {
			if _, err := tea.NewProgram(common.NewRequirements()).Run(); err != nil {
				fmt.Println("Error running program:", err)
				os.Exit(1)
			}
		}

		user := viper.GetString("user")
		operation := viper.GetString("operation")
		group := viper.GetString("group")
		policy := viper.GetString("policy")
		resource := viper.GetString("resource")
		service := viper.GetString("service")
		policyOption := viper.GetString("policy-option")
		terraform := viper.GetBool("terraform")

		// Load the AWS configuration
		awscfg, err := awsconfig.LoadDefaultConfig(context.Background())
		if err != nil {
			return err
		}

		api := internalaws.NewApi(awscfg)
		state := &pkgusers.State{}

		if user != "" {
			awsuser, err := api.FindUser(context.Background(), user)
			if err != nil {
				return err
			}
			state.SetUser(&models.User{
				Name: *awsuser.User.UserName,
				Arn:  *awsuser.User.Arn,
			})
		}

		if operation != "" {
			// Check if the operation exists in the ReachableOperations map
			op, exists := pkgusers.ReachableOperations[pkgusers.OperationType(operation)]
			if !exists {
				return fmt.Errorf("Operation '%s' does not exist in ReachableOperations\n", operation)
			}

			state.SetOperation(&op)
		}

		if policy != "" {
			awspolicy, err := api.FindPolicy(context.Background(), policy)
			if err != nil {
				return err
			}

			state.SetPolicy(&models.Policy{
				Name: *awspolicy.Policy.PolicyName,
				Arn:  *awspolicy.Policy.Arn,
			})
		}

		if group != "" {
			awsgroup, err := api.FindGroup(context.Background(), group)
			if err != nil {
				return err
			}

			state.SetGroup(&models.Group{
				Name: *awsgroup.Group.GroupName,
				Arn:  *awsgroup.Group.Arn,
			})
		}

		if service != "" {
			state.SetService(&models.Service{
				Name: service,
			})
		}

		if resource != "" {
			resourceName := parseResourceNameFromArn(resource)
			state.SetResource(&models.Resource{
				Name: resourceName,
				Arn:  resource,
			})
		}

		if policyOption != "" {
			op, exists := pkgusers.ReachablePolicyOptions[pkgusers.PolicyOptionType(policyOption)]
			if !exists {
				return fmt.Errorf("Policy options '%s' does not exist in ReachableCustomPolicyOptions\n", policyOption)
			}

			state.SetPolicyOption(&op)
		}

		if terraform {
			state.SetTerraform(true)
		}

		controller := pkgusers.NewController(api, cfg.OpenaiApiKey, state)

		p := tea.NewProgram(RootModel(controller.Next()), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}

		return nil
	}
}
