# MCP Code Execution Server - Design Document

## Overview

This PoC implements an MCP server that enables efficient code execution with MCP tools, following the patterns described in Anthropic's "Code execution with MCP" article. The server acts as a bridge between AI agents and multiple MCP servers, allowing agents to write Python scripts that interact with MCP tools more efficiently than direct tool calls.

## Architecture

```
┌─────────────────┐    ┌──────────────────────┐    ┌─────────────────┐
│   AI Agent      │◄──►│  Code Execution      │◄──►│  MCP Server 1   │
│   (STDIO)       │    │  MCP Server          │    │  (HTTP)         │
└─────────────────┘    │  (STDIO)             │    └─────────────────┘
                       │                      │    ┌─────────────────┐
                       │  ┌─────────────────┐ │◄──►│  MCP Server 2   │
                       │  │ Python Sandbox  │ │    │  (HTTP)         │
                       │  │ Environment     │ │    └─────────────────┘
                       │  └─────────────────┘ │    ┌─────────────────┐
                       │                      │◄──►│  MCP Server N   │
                       │  ┌─────────────────┐ │    │  (HTTP)         │
                       │  │ Dynamic Module  │ │    └─────────────────┘
                       │  │ Generator       │ │
                       │  └─────────────────┘ │
                       └──────────────────────┘
```

## Core Components

### 1. MCP Server (Go)
- **Framework**: Official Go MCP SDK
- **Transport**: STDIO (for AI agent connection)
- **Backend Communication**: HTTP clients to MCP servers
- **Configuration**: JSON-based server configuration
- **Tools Exposed**:
  - `search_tools`: Find available tools with filtering
  - `execute_code`: Run Python scripts with MCP tool access

### 2. Python Execution Environment
- **Isolation**: Goroutine-based sandbox with dedicated workspace
- **Limits**: Hierarchical timeouts and memory limits
- **Workspace**: Session-based (cleared after execution)
- **Module Access**: Dynamically generated Python modules for each MCP server

### 3. Dynamic Module Generator
- **Pattern**: Filesystem-based approach (`servers/google_drive/get_document.py`)
- **Generation**: On-demand module creation from MCP tool schemas
- **Storage**: Central location accessible by sandbox processes
- **Updates**: Progressive disclosure - modules generated as needed

## Tool Interface Design

### search_tools Tool
```json
{
  "name": "search_tools",
  "description": "Search available MCP tools with filtering options",
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": {
        "type": "string",
        "description": "Search term for tool names and descriptions"
      },
      "server_filter": {
        "type": "string", 
        "description": "Filter by server name (exact or partial match)"
      },
      "detail_level": {
        "type": "string",
        "enum": ["name", "description", "full"],
        "default": "description",
        "description": "Level of detail to return"
      }
    }
  }
}
```

### execute_code Tool
```json
{
  "name": "execute_code",
  "description": "Execute Python code with access to MCP tools",
  "inputSchema": {
    "type": "object",
    "properties": {
      "code": {
        "type": "string",
        "description": "Python code to execute"
      },
      "timeout": {
        "type": "number",
        "description": "Execution timeout in seconds",
        "default": 30
      }
    },
    "required": ["code"]
  }
}
```

## Python Module Structure

### Generated Module Example
```python
# servers/google_drive/get_document.py
from ..._mcp_client import call_mcp_tool

def get_document(document_id: str, fields: str = None):
    """Retrieves a document from Google Drive
    
    Args:
        document_id: The ID of the document to retrieve
        fields: Specific fields to return
        
    Returns:
        Document object with title, body content, metadata, permissions, etc.
    """
    params = {"document_id": document_id}
    if fields is not None:
        params["fields"] = fields
    
    return call_mcp_tool("google_drive", "get_document", params)
```

### Usage in Agent Scripts
```python
# Agent-generated script
from servers.google_drive import get_document
from servers.salesforce import update_record

# Get document content
doc = get_document("abc123")
content = doc["content"]

# Update Salesforce record
result = update_record(
    object_type="Lead",
    record_id="00Q5f000001abcXYZ", 
    data={"Notes": content}
)

print(f"Updated record: {result['id']}")
```

## Configuration

### Server Configuration (config.json)
```json
{
  "mcp_servers": [
    {
      "name": "google_drive",
      "url": "http://localhost:8001",
      "timeout": 30,
      "expose_all": true
    },
    {
      "name": "salesforce", 
      "url": "http://localhost:8002",
      "timeout": 30,
      "expose_all": true
    }
  ],
  "execution": {
    "default_timeout": 30,
    "max_memory_mb": 512,
    "workspace_path": "./workspace"
  },
  "logging": {
    "level": "info",
    "s3_enabled": true,
    "stdout_enabled": false
  }
}
```

## Implementation Plan

### Phase 1: Core Infrastructure
1. **MCP Server Setup** ✅ COMPLETED
   - ✅ Initialize Go MCP server with STDIO transport
   - ✅ Implement basic tool registration
   - ✅ Add JSON configuration loading

2. **HTTP Client Integration**
   - Create HTTP clients for backend MCP servers
   - Implement MCP-over-HTTP protocol handling
   - Add connection pooling and retry logic

3. **Python Sandbox**
   - Create isolated goroutine execution environment
   - Implement workspace management (session-based)
   - Add basic timeout and resource limits

### Phase 2: Dynamic Module Generation
1. **Tool Discovery**
   - Query backend MCP servers via HTTP for available tools
   - Cache tool schemas and metadata
   - Handle server availability and errors

2. **Module Generator**
   - Create Python module templates
   - Generate modules from MCP tool schemas
   - Implement progressive loading (on-demand)

3. **Tool Integration**
   - Implement `_mcp_client.call_mcp_tool()` function
   - Handle HTTP requests to backend servers
   - Add proper error handling and retries

### Phase 3: Search and Execution
1. **search_tools Implementation**
   - Tool name and description search (exact/partial)
   - Server filtering
   - Detail level support (name/description/full)

2. **execute_code Implementation**
   - Python script execution in sandbox
   - Error handling and logging
   - Return script output or error details

### Phase 4: Monitoring and Efficiency
1. **Token Usage Tracking**
   - Implement context efficiency monitoring
   - Compare direct tool calls vs code execution
   - Generate efficiency reports

2. **Logging Integration**
   - CDL logging SDK integration
   - Debug logging with minimal production logs
   - Error tracking for script generation issues

## HTTP Communication

### Backend Server Discovery
- HTTP GET to `/mcp/tools` endpoint for tool discovery
- Periodic refresh of tool schemas
- Health checks for server availability

### Tool Execution
- HTTP POST to `/mcp/call` endpoint with tool name and parameters
- Proper error handling for HTTP failures
- Timeout management per server configuration

## Security Considerations

### Input Validation
- Validate Python code syntax before execution
- Sanitize configuration inputs
- Validate MCP tool parameters

### Access Controls
- Restrict file system access to dedicated workspace
- Limit HTTP access to configured MCP servers only
- Prevent access to system resources

### Sandbox Isolation
- Goroutine-based isolation with resource limits
- Dedicated workspace per execution session
- No persistent state between executions

## Testing Strategy

### Manual Testing (MVP)
- Test with sample HTTP MCP servers
- Verify tool discovery and module generation
- Validate Python script execution
- Test error handling scenarios

### Future Testing (Post-PoC)
- Unit tests for core functionality
- Integration tests with mocked HTTP MCP servers
- Performance benchmarking vs direct tool calls

## Success Metrics

1. **Functionality**
   - Successfully connect to multiple HTTP MCP servers
   - Generate working Python modules dynamically
   - Execute agent-written scripts successfully

2. **Efficiency**
   - Demonstrate token usage reduction (target: >90% like Anthropic's example)
   - Measure execution time improvements
   - Track context window utilization

3. **Usability**
   - Agent can discover tools via search
   - Agent can write working Python scripts
   - Clear error messages for debugging

## Future Enhancements

- Persistent workspace and skills library
- Multiple language support (JavaScript, Go)
- Container-based isolation
- Advanced resource monitoring
- Production-ready security features
- WebSocket support for real-time MCP servers
