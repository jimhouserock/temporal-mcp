# ⏰🧠 Temporal-MCP Server (HTTP Transport Fork)

**Fork Maintainer:** Jimmy Lin
**Original Project:** [Mocksi/temporal-mcp](https://github.com/Mocksi/temporal-mcp)

An MCP server that bridges AI assistants (like Claude) and Temporal workflows, enabling chat-driven orchestration of complex backend processes.

## 🚀 **What's New in This Fork**

This fork extends the original [Mocksi/temporal-mcp](https://github.com/Mocksi/temporal-mcp) with **HTTP transport support** for web deployment platforms like [Smithery](https://smithery.ai/), while preserving full compatibility with the original stdio transport.

### **Key Enhancements:**
- ✅ **HTTP Transport** — Deploy to web platforms (Smithery, Docker, cloud)
- ✅ **Dual Compatibility** — Works with both web platforms and Claude Desktop
- ✅ **Container Ready** — Includes `Dockerfile` and `smithery.yaml`
- ✅ **CORS Support** — Built-in web browser compatibility
- ✅ **Port Configuration** — Configurable via `PORT` environment variable

## ✨ Key Features

- **🔍 Workflow Discovery** — AI assistants can explore and understand available workflows
- **💬 Natural Language Control** — Trigger complex processes through simple chat commands
- **📊 Real-time Monitoring** — Track workflow progress and status
- **⚡ Smart Caching** — Optimized performance for instant responses
- **🌐 Flexible Deployment** — Works with Claude Desktop or web platforms

## 🏁 Quick Start

### Prerequisites
- **Go 1.21+** and **Temporal Server** ([setup guide](https://docs.temporal.io/docs/server/quick-install/))

### Deployment Options
- **🌐 Web Deployment** — HTTP transport for Smithery, Docker, cloud platforms
- **🖥️ Claude Desktop** — Original stdio transport for local integration

### Installation

#### Web Deployment (Smithery/Docker) 🌐

```bash
git clone https://github.com/jimhouserock/temporal-mcp.git
cd temporal-mcp
```

**Deploy to Smithery:** Push to GitHub → Connect to [Smithery](https://smithery.ai/) → Deploy with `smithery.yaml`

**Or run with Docker:**
```bash
docker build -t temporal-mcp .
docker run -p 8081:8081 -e PORT=8081 temporal-mcp
```

#### Claude Desktop (Original) 🖥️

1. **Setup:** Start your Temporal server and workers ([Money Transfer Demo](https://github.com/temporal-sa/money-transfer-demo/tree/main) recommended)

2. **Build:**
```bash
git clone https://github.com/jimhouserock/temporal-mcp.git
cd temporal-mcp
make build
```

3. **Configure:** Edit `config.yml` with your workflows (see `config.sample.yml` for examples)

4. **Setup Claude:**
```bash
cd examples
./generate_claude_config.sh
cp examples/claude_config.json ~/Library/Application\ Support/Claude/claude_desktop_config.json
```

## 💬 Usage Examples

Talk to your workflows through natural language:

- *"Transfer $100 from account ABC123 to account XYZ789"*
- *"What transfer scenarios are available?"*
- *"Has the workflow completed yet?"*

Temporal MCP translates natural language into properly formatted workflow executions.

## 🛠️ Development

### Common Commands
- `make build` — Build binary
- `make test` — Run tests
- `make run` — Build and run server

## 🔍 Troubleshooting

- **Connection Issues:** Check Temporal server status and `hostPort` in config
- **Workflow Not Found:** Verify workflow registration and namespace settings
- **Claude Integration:** Ensure `claude_config.json` is in correct location and restart Claude

## ⚙️ Configuration

Configure your workflows in `config.yml` with three main sections:
1. **Temporal Connection** — Server connection details
2. **Cache Settings** — Performance optimization
3. **Workflow Definitions** — AI-discoverable workflows

See `config.sample.yml` for a complete example with the [Temporal Money Transfer Demo](https://github.com/temporal-sa/money-transfer-demo).

## 🙏 Attribution

This fork is based on the excellent work by the original [Mocksi/temporal-mcp](https://github.com/Mocksi/temporal-mcp) project. All core functionality, workflow orchestration, and MCP integration concepts are credited to the original authors.

**Fork Changes:** Added HTTP transport support for web deployment platforms while maintaining compatibility with the original stdio transport.

## 📚 Documentation

For detailed configuration examples, best practices, and advanced usage, see the original project documentation at [Mocksi/temporal-mcp](https://github.com/Mocksi/temporal-mcp).

## 📜 License

MIT License - see LICENSE file for details. Contributions welcome!
