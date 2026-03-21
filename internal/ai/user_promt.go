package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type GPTResponse struct {
	Action                string            `json:"action"`
	Principal             map[string]string `json:"principal"`
	RequestedResourceType string            `json:"requested_resource_type"`
	RequestedResource     string            `json:"requested_resource"`
	IsManagedPolicy       bool              `json:"is_managed_policy"`
	Policy                string            `json:"policy"`
	Error                 bool              `json:"error"`
	Confidence            int               `json:"confidence"`
	TerraformOnly         bool              `json:"terraform_only"`
}

var UserPromptSchema = map[string]interface{}{
	"name": "iam_request",
	"schema": map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"action": map[string]interface{}{
				"type":        []string{"string", "null"},
				"description": "The action type. If undetermined, return null.",
				"enum": []string{
					"attach_policy",
					"detach_policy",
					"add_to_group",
					"remove_from_group",
					"attach_custom_policy",
				},
			},
			"error": map[string]interface{}{
				"type":        "boolean",
				"description": "Indicates if the command cannot be managed by the specified actions.",
			},
			"principal": map[string]interface{}{
				"type":        "object",
				"description": "The target entity to which the policy will be attached.",
				"properties": map[string]interface{}{
					"type": map[string]interface{}{
						"type": []string{"string", "null"},
						"enum": []string{"users", "groups", "roles"},
					},
					"name": map[string]interface{}{
						"type":        []string{"string", "null"},
						"description": "The name of the target entity.",
					},
				},
				"required":             []string{"type", "name"},
				"additionalProperties": false,
			},
			"requested_resource_type": map[string]interface{}{
				"type":        []string{"string", "null"},
				"description": "The type of aws resource user wants access for. AWS::S3::Bucket, AWS::RDS::DBCluster, AWS::RDS::DBInstance, AWS::EC2::Instance etc",
			},
			"requested_resource": map[string]interface{}{
				"type":        []string{"string", "null"},
				"description": "The name of the resource user wants access for.",
			},
			"is_managed_policy": map[string]interface{}{
				"type":        "boolean",
				"description": "Indicates if the provided policy is an exact AWS managed policy.",
			},
			"policy": map[string]interface{}{
				"type":        []string{"string", "null"},
				"description": "The name of the policy. If it's too vague, return null. If the user input does not provide any meaningful context, the model must not guess a policy. For Managed policies, use the arn foe example: arn:aws:iam::aws:policy/AdministratorAccess.",
			},
			"confidence": map[string]interface{}{
				"type":        "integer",
				"description": "Confidence level from 1 to 10 about the policy name.",
			},
			"terraform_only": map[string]interface{}{
				"type":        "boolean",
				"description": "Indicates if the user only wants a Terraform snippet.",
			},
		},
		"required": []string{
			"error",
		},
		"additionalProperties": false,
	},
}

func UserPrompt(apiKey, prompt string) (GPTResponse, error) {
	url := "https://api.openai.com/v1/chat/completions"
	schema := UserPromptSchema
	payload := map[string]interface{}{
		"model":       "gpt-4o",
		"temperature": 0.1,
		"messages": []map[string]string{
			{"role": "system", "content": `
				You are an assistant designed to interpret IAM-related requests and convert them into structured JSON objects.
	
				Your task is to:
				1. Analyze the user's input.
				2. Only provide a field if you are very certain (confidence >= 8). If you are not sure, return null for that field.
				3. Identify the requested action, target entity, and resource details. 
				4. If input is vauge return null for specified section.
				5. If target is a specific resource use custom policies. "Example: For 'Allow access to bucket production-data', identify the required resource-specific policy and set isManagedPolicy = false.
				6. Identify policy. If target is a service try getting aws managed policies first. Be certain about aws managed policy names if its wrong correct it.
				7. If the identified policy name has low confidence, set confidence < 5."
				8. If the user input does not provide any meaningful context, the model must not guess a policy.
			`},
			{"role": "user", "content": prompt},
		},
		"response_format": map[string]interface{}{
			"type":        "json_schema",
			"json_schema": schema,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return GPTResponse{}, fmt.Errorf("failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return GPTResponse{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return GPTResponse{}, fmt.Errorf("request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return GPTResponse{}, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var intermediateResponse struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	err = json.Unmarshal(bodyBytes, &intermediateResponse)
	if err != nil {
		return GPTResponse{}, fmt.Errorf("failed to parse intermediate response: %w", err)
	}

	if len(intermediateResponse.Choices) == 0 || intermediateResponse.Choices[0].Message.Content == "" {
		return GPTResponse{}, fmt.Errorf("no content found in response")
	}

	content := intermediateResponse.Choices[0].Message.Content
	content = strings.TrimSpace(content)

	var gptResponse GPTResponse
	err = json.Unmarshal([]byte(content), &gptResponse)
	if err != nil {
		return GPTResponse{}, fmt.Errorf("failed to parse structured content: %w", err)
	}

	return gptResponse, nil
}

func GenerateCLICommand(response GPTResponse) string {
	var flags []string

	if response.Principal != nil {
		if val, ok := response.Principal["type"]; ok && val != "" {
			flags = append(flags, val)
			if val == "users" {
				flags = append(flags, "--user")
			} else if val == "groups" {
				flags = append(flags, "--group")
			} else if val == "roles" {
				flags = append(flags, "--role")
			}
		}
		if val, ok := response.Principal["name"]; ok && val != "" {
			flags = append(flags, val)
		}
	}

	if response.Action != "" {
		flags = append(flags, fmt.Sprintf("--operation %s", response.Action))
	}

	if response.Policy != "" {
		flags = append(flags, fmt.Sprintf("--policy %s", response.Policy))
	}

	if response.RequestedResource != "" {
		flags = append(flags, fmt.Sprintf("--resource %s", response.RequestedResource))
	}

	if response.RequestedResourceType != "" {
		flags = append(flags, fmt.Sprintf("--service %s", response.RequestedResourceType))
	}

	if response.TerraformOnly {
		flags = append(flags, fmt.Sprintf("--terraform %t", response.TerraformOnly))
	}

	if len(flags) == 0 {
		return "No valid flags generated from GPT response."
	}

	return fmt.Sprintf("aws %s", strings.Join(flags, " "))
}
