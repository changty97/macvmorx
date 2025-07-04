macvmorx: macOS Virtual Machine Orchestrator
macvmorx is a lightweight, Kubernetes-like orchestrator designed to manage a fleet of macOS virtual machines running on Mac Mini hardware in a lab environment. It provides centralized monitoring of Mac Mini nodes, facilitates VM provisioning for ephemeral workloads like GitHub Actions runners, and offers a web interface for easy oversight.
🌟 Features
Node Health Monitoring: Receives heartbeats from Mac Mini agents, providing real-time insights into CPU, memory, disk usage, and VM counts.
Scalable Heartbeat Processing: Designed to efficiently handle heartbeats from 100+ Mac Mini systems.
VM Lifecycle Management: Orchestrates the creation and deletion of ephemeral macOS virtual machines.
Intelligent VM Scheduling: Prioritizes Mac Mini nodes that already have the required VM image cached, falling back to nodes capable of downloading it.
Web-based Dashboard: A user-friendly interface to visualize the status of all Mac Mini nodes and their running VMs.
Configurable Refresh Rates: The web dashboard allows selecting refresh intervals (5s, 30s, 60s, or manual).
Efficient VM Image Management: Works in conjunction with the macvmorx-agent to manage large macOS VM images (DMG/IPSW) via a caching and LRU eviction strategy from GCP Cloud Storage.
Command-Line Interface (CLI): Basic CLI for starting the server and future management tasks.
🏗️ Architecture Overview
macvmorx consists of two main components:
macvmorx (Orchestrator - This Repository): The central server application written in Go. It receives heartbeats, maintains node state, implements scheduling logic, and serves the web dashboard and API.
macvmorx-agent (Agent - Separate Repository): A Go application (or potentially Swift for deeper integration) that runs on each Mac Mini. It collects system metrics, manages local VM images (downloading from GCP, LRU caching), provisions/deletes VMs using macOS virtualization tools, and sends heartbeats to the orchestrator.
+-------------------+       +-------------------+       +-------------------+
|   GitHub Actions  |       |   macvmorx Agent  |       |   macvmorx Agent  |
|     Workflow      |       |  (Mac Mini #1)    |       |  (Mac Mini #N)    |
|                   |       |                   |       |                   |
|  - Request VM     | <---> | - Send Heartbeats | <---> | - Send Heartbeats |
|  - Job Completion |       | - Manage VMs      |       | - Manage VMs      |
|                   |       | - Cache Images    |       | - Cache Images    |
+---------^---------+       +---------^---------+       +---------^---------+
          |                           |                           |
          |                           |                           |
          |                           |                           |
          |                           V                           V
          |             +-------------------------------------------------+
          |             |             macvmorx Orchestrator             |
          |             |        (Go Application - This Repo)           |
          |             |                                                 |
          |             |  - Receives Heartbeats                          |
          |             |  - Node State Management (In-memory/DB)         |
          |             |  - VM Scheduling Logic                           |
          |             |  - REST API for Agents & Clients                 |
          |             |  - Serves Web Dashboard                          |
          |             +-------------------------------------------------+
          |                           ^
          |                           |
          +---------------------------+
          (VM Provisioning/Deletion Commands)

+-----------------------------------------------------------------------------+
|                            GCP Cloud Storage                                |
|                        (Central VM Image Repository)                        |
+-----------------------------------------------------------------------------+
          ^
          |
          +-----------------------------------------------------------------+
          (Image Downloads by macvmorx Agents)


🚀 Getting Started
Prerequisites
Go 1.22+ installed
Git
(Optional but Recommended) A running macvmorx-agent on at least one Mac Mini to see heartbeats.
(For Image Management) A GCP Cloud Storage bucket configured with your VM images (DMG/IPSW files).
Installation
Clone the repository:
git clone https://github.com/your-username/macvmorx.git
cd macvmorx


Initialize Go modules:
go mod tidy


Build the orchestrator executable:
go build -o macvmorx ./cmd/macvmorx


Configuration
macvmorx can be configured using environment variables or command-line flags. Command-line flags take precedence.
Environment Variable
Flag
Default Value
Description
MACVMORX_WEB_PORT
--port / -p
8080
Port for the web server and API.
MACVMORX_OFFLINE_TIMEOUT
--offline-timeout
30s
Duration after which a node is considered offline if no heartbeat is received. (e.g., 45s, 1m)
MACVMORX_MONITOR_INTERVAL
--monitor-interval
5s
Interval for the orchestrator to check for offline nodes.

Example using environment variables:
MACVMORX_WEB_PORT=9000 MACVMORX_OFFLINE_TIMEOUT=45s ./macvmorx server


Example using command-line flags:
./macvmorx server --port 9000 --offline-timeout 45s


Running the Orchestrator
To start the macvmorx web server and API:
./macvmorx server


The orchestrator will start listening on the configured port (default: 8080).
Accessing the Web Interface
Once the orchestrator is running, open your web browser and navigate to:
http://localhost:8080 (or your configured port)
You will see a dashboard displaying the status of connected Mac Mini nodes.
Simulating Heartbeats (for testing)
You can send test heartbeats to the orchestrator using curl to see nodes appear on the dashboard:
# Example Heartbeat from mac-mini-001
curl -X POST -H "Content-Type: application/json" -d '{
    "nodeId": "mac-mini-001",
    "vmCount": 1,
    "vms": [
        {"vmId": "vm-001-job-abc", "imageName": "macos-sonoma-runner", "runtimeSeconds": 3600, "vmHostname": "runner-abc", "vmIpAddress": "10.0.0.101"}
    ],
    "cpuUsagePercent": 45.2,
    "memoryUsageGB": 8.5,
    "totalMemoryGB": 16.0,
    "diskUsageGB": 150.0,
    "totalDiskGB": 500.0,
    "status": "healthy",
    "cachedImages": ["macos-sonoma-runner", "macos-ventura-base"]
}' http://localhost:8080/api/heartbeat

# Example Heartbeat from mac-mini-002
curl -X POST -H "Content-Type: application/json" -d '{
    "nodeId": "mac-mini-002",
    "vmCount": 0,
    "vms": [],
    "cpuUsagePercent": 10.5,
    "memoryUsageGB": 4.0,
    "totalMemoryGB": 16.0,
    "diskUsageGB": 100.0,
    "totalDiskGB": 500.0,
    "status": "idle",
    "cachedImages": ["macos-monterey-base"]
}' http://localhost:8080/api/heartbeat


🔌 API Endpoints
The macvmorx orchestrator exposes the following REST API endpoints:
POST /api/heartbeat:
Description: Receives heartbeat payloads from macvmorx-agent instances.
Request Body: application/json (see HeartbeatPayload in internal/models/models.go)
Response: 200 OK with a JSON message.
GET /api/nodes:
Description: Retrieves the current status of all registered Mac Mini nodes.
Response: 200 OK with a JSON array of NodeStatus objects.
POST /api/schedule-vm:
Description: Requests the orchestrator to schedule a new VM on an available Mac Mini.
Request Body: application/json (see VMRequest in internal/models/models.go)
{
    "imageName": "macos-sonoma-github-runner"
}


Response: 200 OK with {"message": "VM scheduled successfully", "nodeId": "..."} or 503 Service Unavailable if no suitable node is found.
🖥️ Web Interface
The web interface is a simple single-page application built with HTML, Tailwind CSS, and JavaScript. It provides a visual overview of your Mac Mini lab:
Lists all connected Mac Mini nodes.
Displays their online/offline status, resource usage (CPU, Memory, Disk).
Shows the number of running VMs, details of each VM (ID, image, runtime, hostname, IP).
Lists the VM images cached on each Mac Mini.
Allows selecting a refresh interval for the dashboard data.
🤖 Agent Software (macvmorx-agent)
The macvmorx-agent is a separate Go application designed to run on each Mac Mini. Its responsibilities include:
Collecting system metrics and sending heartbeats to the orchestrator.
Managing the lifecycle of macOS virtual machines (creation, deletion) using macOS virtualization tools (e.g., vm command-line tool, or potentially direct Hypervisor.framework interaction).
Implementing a local LRU cache for VM images, downloading them from GCP Cloud Storage as needed.
Running a post-provisioning script inside new VMs to install GitHub Actions runner software.
Find the macvmorx-agent repository here: [Link to macvmorx-agent repository] (You'll need to create this separate repository).
💾 VM Image Management
macvmorx leverages a robust VM image management strategy:
Centralized Storage (GCP Cloud Storage): All macOS VM images (DMG/IPSW) are stored in a high-performance GCP Cloud Storage bucket.
Local Caching on Agents: Each macvmorx-agent maintains a local cache directory (/var/macvmorx/images_cache by default) for VM images.
LRU Eviction: The agent implements a Least Recently Used (LRU) eviction policy, keeping a configurable maximum number of images (default: 5) and removing older, less used ones to free up disk space.
On-Demand Download & Waiting: When the orchestrator schedules a VM requiring a specific image:
It first tries to find an available Mac Mini that already has that image cached.
If not, it finds an available Mac Mini and instructs it to download the image from GCP. The VM provisioning process on the agent will then wait for the download to complete before spinning up the VM. This ensures the GitHub job starts only when the VM is ready, without blocking the overall orchestrator.
Checksum Verification: (Future enhancement) Agents should verify image integrity using checksums after download.
🤝 Contributing
Contributions are welcome! Please feel free to open issues, submit pull requests, or suggest improvements.
📄 License
This project is licensed under the MIT License - see the LICENSE file for details.
