import http.server
import socketserver
import json
import threading

PORT = 8001

TOOLS = [
    {
        "name": "mock_tool",
        "description": "A mock tool for testing",
        "inputSchema": {
            "type": "object",
            "properties": {
                "message": {"type": "string"}
            }
        }
    }
]

class MockMCPHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        if self.path == "/mcp/tools":
            self.send_response(200)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps(TOOLS).encode())
        elif self.path == "/health":
            self.send_response(200)
            self.end_headers()
        else:
            self.send_response(404)
            self.end_headers()

    def do_POST(self):
        if self.path == "/mcp/call":
            content_length = int(self.headers['Content-Length'])
            post_data = self.rfile.read(content_length)
            request = json.loads(post_data)
            
            tool_name = request.get("tool")
            params = request.get("params", {})
            
            response = {
                "status": "success",
                "tool": tool_name,
                "result": f"Echo: {params.get('message', '')}"
            }
            
            self.send_response(200)
            self.send_header("Content-type", "application/json")
            self.end_headers()
            self.wfile.write(json.dumps(response).encode())
        else:
            self.send_response(404)
            self.end_headers()

def run_server():
    socketserver.TCPServer.allow_reuse_address = True
    with socketserver.TCPServer(("", PORT), MockMCPHandler) as httpd:
        print(f"Mock MCP Server serving at port {PORT}")
        httpd.serve_forever()

if __name__ == "__main__":
    run_server()
