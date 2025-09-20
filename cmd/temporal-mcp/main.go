package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/mocksi/temporal-mcp/internal/sanitize_history_event"
	"google.golang.org/protobuf/encoding/protojson"

	mcp "github.com/metoro-io/mcp-golang"
	mcphttp "github.com/metoro-io/mcp-golang/transport/http"
	"github.com/mocksi/temporal-mcp/internal/config"
	"github.com/mocksi/temporal-mcp/internal/temporal"
	temporal_enums "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

func main() {
	// Parse command line arguments
	configFile := flag.String("config", "config.yml", "Path to configuration file")
	port := flag.String("port", "", "Port to listen on (overrides PORT env var)")
	flag.Parse()

	// Configure logger to write to stderr
	log.SetOutput(os.Stderr)
	log.Println("Starting Temporal MCP HTTP server...")

	// Setup signal handling for graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	// Load configuration
	cfg, err := config.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	log.Printf("Loaded configuration with %d workflows", len(cfg.Workflows))

	// Initialize Temporal client
	var temporalClient client.Client
	var temporalError error

	temporalClient, temporalError = temporal.NewTemporalClient(cfg.Temporal)
	if temporalError != nil {
		log.Printf("WARNING: Failed to connect to Temporal service: %v", temporalError)
		log.Printf("MCP will run in degraded mode - workflow executions will return errors")
	} else {
		defer temporalClient.Close()
		log.Printf("Connected to Temporal service at %s", cfg.Temporal.HostPort)
	}

	// Determine port to listen on
	listenPort := "8081" // Default port for Smithery
	if *port != "" {
		listenPort = *port
	} else if envPort := os.Getenv("PORT"); envPort != "" {
		listenPort = envPort
	}

	// Create HTTP transport for Smithery deployment
	transport := mcphttp.NewHTTPTransport("/mcp")
	transport.WithAddr(":" + listenPort)

	// Create a new MCP server with HTTP transport
	server := mcp.NewServer(transport)

	// Register all workflow tools
	log.Println("Registering workflow tools...")
	err = registerWorkflowTools(server, cfg, temporalClient)
	if err != nil {
		log.Fatalf("Failed to register workflow tools: %v", err)
	}

	// Register get workflow history tool
	err = registerGetWorkflowHistoryTool(server, temporalClient)
	if err != nil {
		log.Fatalf("Failed to register get workflow history tool: %v", err)
	}

	// Register system prompt
	err = registerSystemPrompt(server, cfg)
	if err != nil {
		log.Fatalf("Failed to register system prompt: %v", err)
	}

	// Start the MCP server (this will start the HTTP server internally)
	go func() {
		log.Printf("Temporal MCP HTTP server listening on port %s", listenPort)
		log.Printf("MCP endpoint available at: http://localhost:%s/mcp", listenPort)

		if err := server.Serve(); err != nil {
			log.Printf("MCP server error: %v", err)
		}
	}()

	// Wait for termination signal
	sig := <-sigCh
	log.Printf("Received signal %v, shutting down server...", sig)

	log.Printf("Temporal MCP HTTP server has been stopped.")
}

// registerWorkflowTools registers all workflow definitions as MCP tools
func registerWorkflowTools(server *mcp.Server, cfg *config.Config, tempClient client.Client) error {
	// Register all workflows as tools
	for name, workflow := range cfg.Workflows {
		err := registerWorkflowTool(server, name, workflow, tempClient, cfg)
		if err != nil {
			return fmt.Errorf("failed to register workflow tool %s: %w", name, err)
		}
		log.Printf("Registered workflow tool: %s", name)
	}

	return nil
}

// registerWorkflowTool registers a single workflow as an MCP tool
func registerWorkflowTool(server *mcp.Server, name string, workflow config.WorkflowDef, tempClient client.Client, cfg *config.Config) error {
	// Define the type for workflow parameters based on fields
	type WorkflowParams struct {
		Params     map[string]string `json:"params"`
		ForceRerun bool              `json:"force_rerun"`
	}

	// Build detailed parameter descriptions for tool registration
	paramDescriptions := "\n\n**Parameters:**\n"
	for _, field := range workflow.Input.Fields {
		for fieldName, description := range field {
			isRequired := !strings.Contains(description, "Optional")
			if isRequired {
				paramDescriptions += fmt.Sprintf("- `%s` (required): %s\n", fieldName, description)
			} else {
				paramDescriptions += fmt.Sprintf("- `%s` (optional): %s\n", fieldName, description)
			}
		}
	}

	// Add example usage
	paramDescriptions += "\n**Example Usage:**\n```json\n{\n  \"params\": {\n"
	paramExamples := []string{}
	for _, field := range workflow.Input.Fields {
		for fieldName, _ := range field {
			if strings.Contains(fieldName, "json") {
				paramExamples = append(paramExamples, fmt.Sprintf("    \"%s\": {\"example\": \"value\"}", fieldName))
			} else if strings.Contains(fieldName, "id") {
				paramExamples = append(paramExamples, fmt.Sprintf("    \"%s\": \"example-id-123\"", fieldName))
			} else {
				paramExamples = append(paramExamples, fmt.Sprintf("    \"%s\": \"example value\"", fieldName))
			}
		}
	}
	paramDescriptions += strings.Join(paramExamples, ",\n")
	paramDescriptions += "\n  },\n  \"force_rerun\": false\n}\n```"

	// Create complete extended purpose description
	extendedPurpose := workflow.Purpose + paramDescriptions

	// Register the tool with MCP server
	return server.RegisterTool(name, extendedPurpose, func(args WorkflowParams) (*mcp.ToolResponse, error) {
		// Check if Temporal client is available
		if tempClient == nil {
			log.Printf("Error: Temporal client is not available for workflow: %s", name)
			return mcp.NewToolResponse(mcp.NewTextContent(
				"Error: Temporal service is currently unavailable. Please try again later.",
			)), nil
		}

		// Validate required parameters before execution
		if args.Params == nil {
			return mcp.NewToolResponse(mcp.NewTextContent(
				fmt.Sprintf("Error: No parameters provided for workflow %s. Please provide required parameters.", name),
			)), nil
		}

		// Build list of required parameters
		var requiredParams []string
		for _, field := range workflow.Input.Fields {
			for fieldName, description := range field {
				if !strings.Contains(description, "Optional") {
					requiredParams = append(requiredParams, fieldName)
				}
			}
		}

		// Check for missing required parameters
		var missingParams []string
		for _, param := range requiredParams {
			if _, exists := args.Params[param]; !exists || args.Params[param] == "" {
				missingParams = append(missingParams, param)
			}
		}

		// Return error if any required parameters are missing
		if len(missingParams) > 0 {
			missingParamsList := strings.Join(missingParams, ", ")
			return mcp.NewToolResponse(mcp.NewTextContent(
				fmt.Sprintf("Error: Missing required parameters for workflow %s: %s", name, missingParamsList),
			)), nil
		}

		// Execute the workflow
		// Determine which task queue to use (workflow-specific or default)
		taskQueue := workflow.TaskQueue
		if taskQueue == "" && cfg != nil {
			taskQueue = cfg.Temporal.DefaultTaskQueue
			log.Printf("Using default task queue: %s for workflow %s", taskQueue, name)
		}

		workflowID, err := computeWorkflowID(workflow, args.Params)
		if err != nil {
			log.Printf("Error computing workflow ID from arguments: %v", err)
			return mcp.NewToolResponse(mcp.NewTextContent(
				fmt.Sprintf("Error computing workflow ID from arguments: %v", err),
			)), nil
		}

		if workflowID == "" {
			log.Printf("Workflow %q has an empty or missing workflowIDRecipe - using a random workflow id", name)
			workflowID = uuid.NewString()
		}

		// This will execute a new workflow when:
		// - there is no workflow with the given id
		// - there is a failed workflow with the given id (e.g. terminated, failed, timed out)
		// and attach to an existing workflow when:
		// - there is a running workflow with the given id
		// - there is a successful workflow with the given id
		//
		// Note that temporal's data retention window (a setting on each namespace) influences the behavior above
		reusePolicy := temporal_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY
		conflictPolicy := temporal_enums.WORKFLOW_ID_CONFLICT_POLICY_USE_EXISTING

		if args.ForceRerun {
			// This will execute a new workflow in all cases. If there is a running workflow with the given id, it will
			// be terminated.
			reusePolicy = temporal_enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE
			conflictPolicy = temporal_enums.WORKFLOW_ID_CONFLICT_POLICY_TERMINATE_EXISTING
		}

		wfOptions := client.StartWorkflowOptions{
			TaskQueue:                taskQueue,
			ID:                       workflowID,
			WorkflowIDReusePolicy:    reusePolicy,
			WorkflowIDConflictPolicy: conflictPolicy,
		}

		log.Printf("Starting workflow %s on task queue %s", name, taskQueue)

		// Start workflow execution
		run, err := tempClient.ExecuteWorkflow(context.Background(), wfOptions, name, args.Params)
		if err != nil {
			log.Printf("Error starting workflow %s: %v", name, err)
			return mcp.NewToolResponse(mcp.NewTextContent(
				fmt.Sprintf("Error executing workflow: %v", err),
			)), nil
		}

		log.Printf("Workflow started: WorkflowID=%s RunID=%s", run.GetID(), run.GetRunID())

		// Wait for workflow completion
		var result string
		if err := run.Get(context.Background(), &result); err != nil {
			log.Printf("Error in workflow %s execution: %v", name, err)
			return mcp.NewToolResponse(mcp.NewTextContent(
				fmt.Sprintf("Workflow failed: %v", err),
			)), nil
		}

		log.Printf("Workflow %s completed successfully", name)

		return mcp.NewToolResponse(mcp.NewTextContent(result)), nil
	})
}

func computeWorkflowID(workflow config.WorkflowDef, params map[string]string) (string, error) {
	tmpl := template.New("id_recipe")

	tmpl.Funcs(template.FuncMap{
		"hash": func(paramsToHash ...any) (string, error) {
			return hashWorkflowArgs(params, paramsToHash...)
		},
	})
	if _, err := tmpl.Parse(workflow.WorkflowIDRecipe); err != nil {
		return "", err
	}

	writer := strings.Builder{}
	if err := tmpl.Execute(&writer, params); err != nil {
		return "", err
	}

	return writer.String(), nil
}

// registerGetWorkflowHistoryTool registres a tool that gets workflow histories
func registerGetWorkflowHistoryTool(server *mcp.Server, tempClient client.Client) error {
	type GetWorkflowHistoryParams struct {
		WorkflowID string `json:"workflowId"`
		RunID      string `json:"runId"`
	}
	desc := "Gets the workflow execution history for a specific run of a workflow. runId is optional - if omitted, this tool gets the history for the latest run of the given workflowId"

	return server.RegisterTool("GetWorkflowHistory", desc, func(args GetWorkflowHistoryParams) (*mcp.ToolResponse, error) {
		// Check if Temporal client is available
		if tempClient == nil {
			log.Printf("Error: Temporal client is not available for getting workflow histories")
			return mcp.NewToolResponse(mcp.NewTextContent(
				"Error: Temporal client is not available for getting workflow histories",
			)), nil
		}

		eventJsons := make([]string, 0)
		iterator := tempClient.GetWorkflowHistory(context.Background(), args.WorkflowID, args.RunID, false, temporal_enums.HISTORY_EVENT_FILTER_TYPE_ALL_EVENT)
		for iterator.HasNext() {
			event, err := iterator.Next()
			if err != nil {
				msg := fmt.Sprintf("Error: Failed to get %dth history event: %v", len(eventJsons), err)
				log.Print(msg)
				return mcp.NewToolResponse(mcp.NewTextContent(msg)), nil
			}

			sanitize_history_event.SanitizeHistoryEvent(event)
			bytes, err := protojson.Marshal(event)
			if err != nil {
				// should never happen?
				return nil, err
			}

			eventJsons = append(eventJsons, string(bytes))
		}

		// The last step of json-marshalling is unfortunate (forced on us by the lack of a proto for the list of
		// events), but not worth actually building and marshalling a slice for. Let's just do it by hand.
		allEvents := strings.Builder{}
		allEvents.WriteString("[")
		for i, eventJson := range eventJsons {
			if i > 0 {
				allEvents.WriteString(",")
			}
			allEvents.WriteString(eventJson)
		}
		allEvents.WriteString("]")

		return mcp.NewToolResponse(mcp.NewTextContent(allEvents.String())), nil
	})
}

// registerSystemPrompt registers the system prompt for the MCP
func registerSystemPrompt(server *mcp.Server, cfg *config.Config) error {
	return server.RegisterPrompt("system_prompt", "System prompt for the Temporal MCP", func(_ struct{}) (*mcp.PromptResponse, error) {
		// Build list of available tools from workflows
		workflowList := ""
		for name, workflow := range cfg.Workflows {
			// Use the complete purpose which already includes parameter details from config.yml
			detailedPurpose := workflow.Purpose

			workflowList += fmt.Sprintf("## %s\n", name)
			workflowList += fmt.Sprintf("**Purpose:** %s\n\n", detailedPurpose)
			workflowList += fmt.Sprintf("**Input Type:** %s\n\n", workflow.Input.Type)

			// Add parameters section with detailed formatting based on the Input.Fields
			workflowList += "**Parameters:**\n"
			for _, field := range workflow.Input.Fields {
				for fieldName, description := range field {
					isRequired := !strings.Contains(description, "Optional")
					if isRequired {
						workflowList += fmt.Sprintf("- `%s` (required): %s\n", fieldName, description)
					} else {
						workflowList += fmt.Sprintf("- `%s` (optional): %s\n", fieldName, description)
					}
				}
			}

			// Add example of how to call this workflow
			workflowList += "\n**Example Usage:**\n"
			workflowList += "```json\n"
			workflowList += "{\n  \"params\": {\n"

			// Generate example parameters
			paramExamples := []string{}
			for _, field := range workflow.Input.Fields {
				for fieldName, _ := range field {
					if strings.Contains(fieldName, "json") {
						paramExamples = append(paramExamples, fmt.Sprintf("    \"%s\": {\"example\": \"value\"}", fieldName))
					} else if strings.Contains(fieldName, "id") {
						paramExamples = append(paramExamples, fmt.Sprintf("    \"%s\": \"example-id-123\"", fieldName))
					} else {
						paramExamples = append(paramExamples, fmt.Sprintf("    \"%s\": \"example value\"", fieldName))
					}
				}
			}
			workflowList += strings.Join(paramExamples, ",\n")
			workflowList += "\n  },\n  \"force_rerun\": false\n}\n```\n"

			// Add output information
			workflowList += fmt.Sprintf("\n**Output Type:** %s\n", workflow.Output.Type)
			if workflow.Output.Description != "" {
				workflowList += fmt.Sprintf("**Output Description:** %s\n", workflow.Output.Description)
			}

			// Extract required parameters for validation guidance
			var requiredParams []string
			for _, field := range workflow.Input.Fields {
				for fieldName, description := range field {
					if !strings.Contains(description, "Optional") {
						requiredParams = append(requiredParams, fieldName)
					}
				}
			}

			// Add validation guidelines
			if len(requiredParams) > 0 {
				workflowList += "\n**Required Validation:**\n"
				workflowList += "- Validate all required parameters are provided before execution\n"
				paramsList := strings.Join(requiredParams, ", ")
				workflowList += fmt.Sprintf("- Required parameters: %s\n", paramsList)
			}

			workflowList += "\n---\n\n"
		}

		systemPrompt := fmt.Sprintf(`You are now connected to a Temporal MCP (Model Control Protocol) server that provides access to various Temporal workflows.

This MCP exposes the following workflow tools:

%s
## Parameter Validation Guidelines

Before executing any workflow, ensure you:

1. Validate all required parameters are present and properly formatted
2. Check that string parameters have appropriate length and format
3. Verify numeric parameters are within expected ranges
4. Ensure any IDs follow the proper format guidelines
5. Ask the user for any missing required parameters before execution

## Tool Usage Instructions

Use these tools to help users interact with Temporal workflows. Each workflow requires a 'params' object containing the necessary parameters listed above.

When constructing your calls:
- Include all required parameters
- Set force_rerun to true only when explicitly requested by the user
- When force_rerun is false, Temporal will deduplicate workflows based on their arguments

## General Example Structure

To call any workflow:
`+"```"+`
{
  "params": {
    "param1": "value1",
    "param2": "value2"
  },
  "force_rerun": false
}
`+"```"+`

Refer to each workflow's specific example above for exact parameter requirements.`, workflowList)

		return mcp.NewPromptResponse("system_prompt", mcp.NewPromptMessage(mcp.NewTextContent(systemPrompt), mcp.Role("system"))), nil
	})
}

// hashWorkflowArgs produces a short (suitable for inclusion in workflow id) hash of the given arguments. Args must be
// json.Marshal-able.
func hashWorkflowArgs(allParams map[string]string, paramsToHash ...any) (string, error) {
	if len(paramsToHash) == 0 {
		log.Printf("Warning: No hash arguments provided - will hash all arguments. Please replace {{ hash }} with {{ hash . }} in the workflowIDRecipe")
		paramsToHash = []any{allParams}
	}

	hasher := fnv.New32()
	for _, arg := range paramsToHash {
		// important: json.Marshal sorts map keys
		bytes, err := json.Marshal(arg)
		if err != nil {
			return "", err
		}
		_, _ = hasher.Write(bytes)
	}
	return fmt.Sprintf("%d", hasher.Sum32()), nil
}
