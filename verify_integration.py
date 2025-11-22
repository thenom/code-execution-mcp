import subprocess
import time
import json
import os
import sys

def run_verification():
    # 1. Start Mock Server
    mock_server = subprocess.Popen(["python3", "mock_mcp_server.py"])
    print("Started Mock Server")
    time.sleep(2) # Wait for server to start

    try:
        # 2. Build the Go server
        print("Building Go Server...")
        subprocess.check_call(["go", "build", "-o", "server", "main.go"])
        
        # 3. Create config for testing
        config = {
            "mcp_servers": [
                {
                    "name": "mock_server",
                    "url": "http://localhost:8001",
                    "timeout": 5,
                    "expose_all": True
                }
            ],
            "execution": {
                "default_timeout": 10,
                "max_memory_mb": 128,
                "workspace_path": "./test_workspace"
            },
            "logging": {
                "level": "debug",
                "s3_enabled": False,
                "stdout_enabled": True
            }
        }
        with open("test_config.json", "w") as f:
            json.dump(config, f)

        # 4. Run the Go server and send a request
        print("Running Go Server...")
        server_process = subprocess.Popen(
            ["./server", "-config", "test_config.json"],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )

        # Construct JSON-RPC request
        # The server expects MCP protocol messages.
        # Since we are using mcp-go/server.ServeStdio, it handles JSON-RPC 2.0
        
        # First, initialize
        init_req = {
            "jsonrpc": "2.0",
            "id": 1,
            "method": "initialize",
            "params": {
                "protocolVersion": "2024-11-05",
                "capabilities": {},
                "clientInfo": {"name": "test-client", "version": "1.0"}
            }
        }
        
        # Send initialize
        server_process.stdin.write(json.dumps(init_req) + "\n")
        server_process.stdin.flush()
        
        # Read initialize response
        print("Sent initialize...")
        while True:
            line = server_process.stdout.readline()
            if not line:
                break
            print(f"Received: {line.strip()}")
            resp = json.loads(line)
            if resp.get("id") == 1:
                break
        
        # Send initialized notification
        init_notif = {
            "jsonrpc": "2.0",
            "method": "notifications/initialized"
        }
        server_process.stdin.write(json.dumps(init_notif) + "\n")
        server_process.stdin.flush()

        # Call execute_code
        # The script should import the mock tool and call it
        python_code = """
from servers.mock_server import mock_tool

result = mock_tool(message="Hello from Python!")
print(result)
"""
        
        call_req = {
            "jsonrpc": "2.0",
            "id": 2,
            "method": "tools/call",
            "params": {
                "name": "execute_code",
                "arguments": {
                    "code": python_code
                }
            }
        }
        
        print("Sending execute_code request...")
        server_process.stdin.write(json.dumps(call_req) + "\n")
        server_process.stdin.flush()
        
        # Read response
        while True:
            line = server_process.stdout.readline()
            if not line:
                break
            print(f"Received: {line.strip()}")
            resp = json.loads(line)
            if resp.get("id") == 2:
                # Check result
                if "error" in resp:
                    print("Error:", resp["error"])
                    sys.exit(1)
                
                content = resp["result"]["content"][0]["text"]
                print("Execution Result:", content)
                
                if "Echo: Hello from Python!" in content:
                    print("VERIFICATION SUCCESSFUL")
                else:
                    print("VERIFICATION FAILED: Unexpected output")
                    sys.exit(1)
                break

    finally:
        # Cleanup
        mock_server.terminate()
        if 'server_process' in locals():
            server_process.terminate()
        
        # Clean up workspace
        # subprocess.call(["rm", "-rf", "test_workspace"])
        # subprocess.call(["rm", "test_config.json", "server"])

if __name__ == "__main__":
    run_verification()
