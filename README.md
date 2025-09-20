# ⏰🧠 Temporal-MCP Server
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/Mocksi/temporal-mcp)
[![CI Status](https://github.com/Mocksi/temporal-mcp/actions/workflows/go-test.yml/badge.svg)](https://github.com/Mocksi/temporal-mcp/actions/workflows/go-test.yml)

**Author:** Jimmy Lin

Temporal MCP is an MCP server that bridges AI assistants (like Claude) and Temporal workflows. It turns complex backend orchestration into simple, chat-driven commands. Imagine triggering stateful processes without writing a line of glue code. Temporal-MCP makes that possible.

## 🚀 **New in This Fork: HTTP Transport Support**

This fork adds **HTTP transport support** for web-based deployment platforms like [Smithery](https://smithery.ai/), while maintaining full compatibility with the original stdio transport for Claude Desktop.

### **Key Differences:**
- ✅ **Dual Transport Support** — Works with both Claude Desktop (stdio) and web platforms (HTTP)
- ✅ **Smithery Ready** — Deploy directly to Smithery with included `smithery.yaml` and `Dockerfile`
- ✅ **CORS Enabled** — Built-in CORS support for web browser compatibility
- ✅ **Container Optimized** — Docker containerization for cloud deployment
- ✅ **Port Flexible** — Configurable via `PORT` environment variable (defaults to 8081)

## Why Temporal MCP

- **Supercharged AI** — AI assistants gain reliable, long-running workflow superpowers
- **Conversational Orchestration** — Trigger, monitor, and manage workflows through natural language
- **Enterprise-Ready** — Leverage Temporal's retries, timeouts, and persistence—exposed in plain text

## ✨ Key Features

- **🔍 Automatic Discovery** — Explore available workflows and see rich metadata
- **🏃‍♂️ Seamless Execution** — Kick off complex processes with a single chat message
- **📊 Real-time Monitoring** — Follow progress, check status, and get live updates
- **⚡ Performance Optimization** — Smart caching for instant answers
- **🧠 AI-Friendly Descriptions** — Purpose fields written for both humans and machines

## 🏁 Getting Started

### Prerequisites

- **Go 1.21+** — For building and running the MCP server
- **Temporal Server** — Running locally or remotely (see [Temporal docs](https://docs.temporal.io/docs/server/quick-install/))

### Deployment Options

**🖥️ Claude Desktop (Original)** — Use stdio transport for local Claude Desktop integration
**🌐 Web Deployment (New)** — Use HTTP transport for Smithery, Docker, or other web platforms

### Quick Install

#### Option 1: Web Deployment (Smithery/Docker) 🌐

1. **Clone and configure**
```bash
git clone https://github.com/jimhouserock/temporal-mcp.git
cd temporal-mcp
```

2. **Deploy to Smithery**
   - Push to GitHub
   - Connect to [Smithery](https://smithery.ai/)
   - Deploy using the included `smithery.yaml`

3. **Or run with Docker**
```bash
docker build -t temporal-mcp .
docker run -p 8081:8081 -e PORT=8081 temporal-mcp
```

#### Option 2: Claude Desktop (Original) 🖥️

1. **Run your Temporal server and workers**
In this example, we'll use the [Temporal Money Transfer Demo](https://github.com/temporal-sa/money-transfer-demo/tree/main).

2. **Build the server**
```bash
git clone https://github.com/jimhouserock/temporal-mcp.git
cd temporal-mcp
make build
```

3. **Define your workflows** in `config.yml`
The sample configuration (`config.sample.yml`) is designed to work with the [Temporal Money Transfer Demo](https://github.com/temporal-sa/money-transfer-demo/tree/main):

```yaml
workflows:
  AccountTransferWorkflow:
    purpose: "Transfers money between accounts with validation and notification. Handles the happy path scenario where everything works as expected."
    input:
      type: "TransferInput"
      fields:
        - from_account: "Source account ID"
        - to_account: "Destination account ID"
        - amount: "Amount to transfer"
    output:
      type: "TransferOutput"
      description: "Transfer confirmation with charge ID"
    taskQueue: "account-transfer-queue"

  AccountTransferWorkflowScenarios:
    purpose: "Extended account transfer workflow with various scenarios including human approval, recoverable failures, and advanced visibility features."
    input:
      type: "TransferInput"
      fields:
        - from_account: "Source account ID"
        - to_account: "Destination account ID"
        - amount: "Amount to transfer"
        - scenario_type: "Type of scenario to execute (human_approval, recoverable_failure, advanced_visibility)"
    output:
      type: "TransferOutput"
      description: "Transfer confirmation with charge ID"
    taskQueue: "account-transfer-queue"
```

4. **Generate Claude's configuration**
```bash
cd examples
./generate_claude_config.sh
```

5. **Install the configuration**
```bash
cp examples/claude_config.json ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

6. **Start Claude** with this configuration

### Conversing with Your Workflows

Now for the magic part! Talk to your workflows through Claude using natural language:

> 💬 "Claude, can you transfer $100 from account ABC123 to account XYZ789?"

> 💬 "What transfer scenarios are available to test?"

> 💬 "Execute a transfer that requires human approval for $500 between accounts ABC123 and XYZ789"

> 💬 "Has the transfer workflow completed yet?"

> 💬 "Run a transfer scenario with recoverable failures to test error handling"

Behind the scenes, Temporal MCP translates these natural language requests into properly formatted workflow executions—no more complex API calls or parameter formatting!

## Core Values

1. **Clarity First** — Use simple, direct language. Avoid jargon.
2. **Benefit-Driven** — Lead with "what's in it for me".
3. **Concise Power** — Less is more—keep sentences tight and memorable.
4. **Personality & Voice** — Bold statements, short lines, a dash of excitement.

## Ready to Showcase

Lights, camera, action—capture your first AI-driven workflow in motion. Share that moment. Inspire others to see Temporal MCP in action.

## Development

### Project Structure

```
./
├── cmd/            # Entry points for executables
├── internal/       # Internal package code
│   ├── api/        # MCP API implementation
│   ├── cache/      # Caching layer
│   ├── config/     # Configuration management
│   └── temporal/   # Temporal client integration
├── examples/       # Example configurations and scripts
└── docs/           # Documentation
```

### Common Commands

| Command | Description |
|---------|-------------|
| `make build` | Builds the binary in `./bin` |
| `make test` | Runs all unit tests |
| `make fmt` | Formats code according to Go standards |
| `make run` | Builds and runs the server |
| `make clean` | Removes build artifacts |

## 🔍 Troubleshooting

### Common Issues

**Connection Refused**
- ✓ Check if Temporal server is running
- ✓ Verify hostPort is correct in config.yml

**Workflow Not Found**
- ✓ Ensure workflow is registered in Temporal
- ✓ Check namespace settings in config.yml

**Claude Can't See Workflows**
- ✓ Verify claude_config.json is in the correct location
- ✓ Restart Claude after configuration changes

## ⚙️ Configuration

The heart of Temporal MCP is its configuration file, which connects your AI assistants to your workflow engine:

### Configuration Architecture

Your `config.yml` consists of three key sections:

1. **🔌 Temporal Connection** — How to connect to your Temporal server
2. **💾 Cache Settings** — Performance optimization for workflow results
3. **🔧 Workflow Definitions** — The workflows your AI can discover and use

### Example Configuration

The sample configuration is designed to work with the Temporal Money Transfer Demo:

```yaml
# Temporal server connection details
temporal:
  hostPort: "localhost:7233"       # Your Temporal server address
  namespace: "default"             # Temporal namespace
  environment: "local"             # "local" or "remote"
  defaultTaskQueue: "account-transfer-queue"  # Default task queue for workflows

  # Fine-tune connection behavior
  timeout: "5s"                    # Connection timeout
  retryOptions:                     # Robust retry settings
    initialInterval: "100ms"       # Start with quick retries
    maximumInterval: "10s"         # Max wait between retries
    maximumAttempts: 5              # Don't try forever
    backoffCoefficient: 2.0         # Exponential backoff

# Define AI-discoverable workflows
workflows:
  AccountTransferWorkflow:
    purpose: "Transfers money between accounts with validation and notification. Handles the happy path scenario where everything works as expected."
    workflowIDRecipe: "transfer_{{.from_account}}_{{.to_account}}_{{.amount}}"
    input:
      type: "TransferInput"
      fields:
        - from_account: "Source account ID"
        - to_account: "Destination account ID"
        - amount: "Amount to transfer"
    output:
      type: "TransferOutput"
      description: "Transfer confirmation with charge ID"
    taskQueue: "account-transfer-queue"
    activities:
      - name: "validate"
        timeout: "5s"
      - name: "withdraw"
        timeout: "5s"
      - name: "deposit"
        timeout: "5s"
      - name: "sendNotification"
        timeout: "5s"
      - name: "undoWithdraw"
        timeout: "5s"
```

> 💡 **Pro Tip:** The sample configuration is pre-configured to work with the [Temporal Money Transfer Demo](https://github.com/temporal-sa/money-transfer-demo/tree/main). Use it as a starting point for your own workflows.

## 💎 Best Practices

### Crafting Perfect Purpose Fields

The `purpose` field is your AI assistant's window into understanding what each workflow does. Make it count!

#### ✅ Do This
- Write clear, detailed descriptions of functionality
- Mention key parameters and how they customize behavior
- Describe expected outputs and their formats
- Note any limitations or constraints

#### ❌ Avoid This
- Vague descriptions ("Processes data")
- Technical jargon without explanation
- Missing important parameters
- Ignoring error cases or limitations

#### Before & After

**Before:** "Gets information about a file."

**After:** "Retrieves detailed metadata about a file or directory including size, creation time, last modified time, permissions, and type. Performs access validation to ensure the requested file is within allowed directories. Returns formatted JSON with all attributes or an appropriate error message."

### Naming Conventions

| Item | Convention | Example |
|------|------------|----------|
| Workflow IDs | PascalCase | `AccountTransferWorkflow` |
| Parameter names | snake_case | `from_account`, `to_account` |
| Parameters with units | Include unit | `timeout_seconds`, `amount` |

### Security Guidelines

⚠️ **Important Security Notes:**

- Keep credentials out of your configuration files
- Use environment variables for sensitive values
- Consider access controls for workflows with sensitive data
- Validate and sanitize all workflow inputs

> 💡 **Tip:** Create different configurations for development and production environments

### Why Good Purpose Fields Matter

1. **Enhanced AI Understanding** - Claude and other AI tools can provide much more accurate and helpful responses when they fully understand the capabilities and limitations of each component
2. **Fewer Errors** - Detailed descriptions reduce the chances of AI systems using components incorrectly
3. **Improved Debugging** - Clear descriptions help identify issues when workflows don't behave as expected
4. **Better Developer Experience** - New team members can understand your system more quickly
5. **Documentation As Code** - Purpose fields serve as living documentation that stays in sync with the codebase

## Contribute & Collaborate

We're building this together.
- Share your own workflow configs
- Improve descriptions
- Share your demos (video or GIF) in issues

Let's unleash the power of AI and Temporal together!

## 📜 License

This project is licensed under the MIT License - see the LICENSE file for details.
Contributions welcome!
