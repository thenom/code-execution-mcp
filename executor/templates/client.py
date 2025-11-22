import json
import urllib.request
import urllib.error

# This dictionary will be populated during generation
SERVER_URLS = {}

def call_mcp_tool(server_name, tool_name, params):
    """
    Calls a tool on a backend MCP server via HTTP.
    
    Args:
        server_name: The name of the server to call
        tool_name: The name of the tool to execute
        params: A dictionary of parameters for the tool
        
    Returns:
        The result of the tool execution
    
    Raises:
        ValueError: If the server is unknown
        RuntimeError: If the HTTP request fails
    """
    if server_name not in SERVER_URLS:
        raise ValueError(f"Unknown server: {server_name}")
        
    base_url = SERVER_URLS[server_name]
    url = f"{base_url}/mcp/call"
    
    payload = {
        "tool": tool_name,
        "params": params
    }
    
    data = json.dumps(payload).encode('utf-8')
    req = urllib.request.Request(url, data=data, headers={'Content-Type': 'application/json'})
    
    try:
        with urllib.request.urlopen(req) as response:
            if response.status != 200:
                raise RuntimeError(f"Server returned status {response.status}")
            
            response_body = response.read().decode('utf-8')
            return json.loads(response_body)
            
    except urllib.error.HTTPError as e:
        error_body = e.read().decode('utf-8')
        raise RuntimeError(f"HTTP Error {e.code}: {error_body}")
    except urllib.error.URLError as e:
        raise RuntimeError(f"Connection failed: {e.reason}")
    except Exception as e:
        raise RuntimeError(f"Execution failed: {str(e)}")
