package aws

import (
	"context"
	"fmt"
	"os"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	internalaws "github.com/Permify/targe/internal/aws"
	"github.com/Permify/targe/internal/config"
	pkggroups "github.com/Permify/targe/pkg/aws/groups"
	"github.com/Permify/targe/pkg/aws/models"
)

type Groups struct {
	model tea.Model
}

func (m Groups) Init() tea.Cmd {
	return m.model.Init() // rest methods are just wrappers for the model's methods
}

func (m Groups) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m.model.Update(msg)
}

func (m Groups) View() string {
	return m.model.View()
}

// NewGroupsCommand -
func NewGroupsCommand(cfg *config.Config) *cobra.Command {
	command := &cobra.Command{
		Use:   "groups",
		Short: "",
		RunE:  groups(cfg),
	}

	f := command.Flags()

	f.String("group", "", "group")
	f.String("operation", "", "operation")
	f.String("policy", "", "policy")
	f.String("resource", "", "resource")
	f.String("service", "", "service")
	f.String("policy-option", "", "policy option")
	f.String("terraform", "", "terraform")

	// SilenceUsage is set to true to suppress usage when an error occurs
	command.SilenceUsage = true

	command.PreRun = func(cmd *cobra.Command, args []string) {
		RegisterGroupsFlags(f)
	}

	return command
}

func groups(cfg *config.Config) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// get min coverage from viper
		group := viper.GetString("group")
		operation := viper.GetString("operation")
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
		state := &pkggroups.State{}

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

		if operation != "" {
			// Check if the operation exists in the ReachableOperations map
			op, exists := pkggroups.ReachableOperations[pkggroups.OperationType(operation)]
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

		if resource != "" {
			resourceName := parseResourceNameFromArn(resource)
			state.SetResource(&models.Resource{
				Name: resourceName,
				Arn:  resource,
			})
		}

		if service != "" {
			state.SetService(&models.Service{
				Name: service,
			})
		}

		if policyOption != "" {
			// Check if the operation exists in the ReachableOperations map
			op, exists := pkggroups.ReachablePolicyOptions[pkggroups.PolicyOptionType(policyOption)]
			if !exists {
				return fmt.Errorf("Policy options '%s' does not exist in ReachableCustomPolicyOptions\n", policyOption)
			}

			state.SetPolicyOption(&op)
		}

		if terraform {
			state.SetTerraform(true)
		}

		controller := pkggroups.NewController(api, cfg.OpenaiApiKey, state)

		p := tea.NewProgram(RootModel(controller.Next()), tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}

		return nil
	}
}
