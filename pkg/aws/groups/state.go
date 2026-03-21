package groups

import (
	"github.com/Permify/targe/pkg/aws/models"
)

// State represents the groups flow state.
type State struct {
	group        *models.Group
	operation    *models.Operation
	policyOption *models.PolicyOption
	service      *models.Service
	resource     *models.Resource
	policy       *models.Policy
	terraform    bool
}

// Getters

// GetGroup retrieves the group from the state.
func (s *State) GetGroup() *models.Group {
	return s.group
}

// GetOperation retrieves the operation from the state.
func (s *State) GetOperation() *models.Operation {
	return s.operation
}

// GetPolicyOption retrieves the policy option from the state.
func (s *State) GetPolicyOption() *models.PolicyOption {
	return s.policyOption
}

// GetService retrieves the service from the state.
func (s *State) GetService() *models.Service {
	return s.service
}

// GetResource retrieves the resource from the state.
func (s *State) GetResource() *models.Resource {
	return s.resource
}

// GetPolicy retrieves the policy from the state.
func (s *State) GetPolicy() *models.Policy {
	return s.policy
}

// Setters

// SetGroup updates the group in the state.
func (s *State) SetGroup(group *models.Group) {
	s.group = group
}

// SetOperation updates the action in the state.
func (s *State) SetOperation(operation *models.Operation) {
	s.operation = operation
}

// SetPolicyOption updates the policy option in the state.
func (s *State) SetPolicyOption(policyOption *models.PolicyOption) {
	s.policyOption = policyOption
}

// SetService updates the service in the state.
func (s *State) SetService(service *models.Service) {
	s.service = service
}

// SetResource updates the resource in the state.
func (s *State) SetResource(resource *models.Resource) {
	s.resource = resource
}

// SetPolicy updates the policy in the state.
func (s *State) SetPolicy(policy *models.Policy) {
	s.policy = policy
}

func (s *State) SetTerraform(terraform bool) {
	s.terraform = terraform
}

func (s *State) GetTerraform() bool {
	return s.terraform
}
