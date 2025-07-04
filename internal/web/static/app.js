document.addEventListener('DOMContentLoaded', () => {
    const nodesContainer = document.getElementById('nodes-container');
    const loadingIndicator = document.getElementById('loading');
    const errorDisplay = document.getElementById('error');

    const fetchNodes = async () => {
        nodesContainer.innerHTML = ''; // Clear previous nodes
        loadingIndicator.classList.remove('hidden');
        errorDisplay.classList.add('hidden');

        try {
            const response = await fetch('/api/nodes');
            if (!response.ok) {
                throw new Error(`HTTP error! status: ${response.status}`);
            }
            const nodes = await response.json();
            displayNodes(nodes);
        } catch (error) {
            console.error('Failed to fetch nodes:', error);
            errorDisplay.classList.remove('hidden');
        } finally {
            loadingIndicator.classList.add('hidden');
        }
    };

    const displayNodes = (nodes) => {
        if (nodes.length === 0) {
            nodesContainer.innerHTML = '<p class="text-center text-gray-600 col-span-full">No Mac Mini nodes registered yet. Waiting for heartbeats...</p>';
            return;
        }

        nodes.sort((a, b) => a.nodeId.localeCompare(b.nodeId)); // Sort by NodeID

        nodes.forEach(node => {
            const statusColor = node.isOnline ? 'bg-green-500' : 'bg-red-500';
            const statusText = node.isOnline ? 'Online' : 'Offline';
            const lastSeen = new Date(node.lastSeen).toLocaleString();

            const nodeCard = `
                <div class="bg-white rounded-lg shadow-lg p-6 flex flex-col justify-between">
                    <div>
                        <h2 class="text-2xl font-semibold text-gray-900 mb-2">${node.nodeId}</h2>
                        <div class="flex items-center mb-4">
                            <span class="w-3 h-3 rounded-full ${statusColor} mr-2"></span>
                            <span class="text-lg font-medium text-gray-700">${statusText}</span>
                        </div>
                        <p class="text-gray-600 mb-2"><strong>VMs Running:</strong> ${node.vmCount}/2</p>
                        <p class="text-gray-600 mb-2"><strong>CPU Usage:</strong> ${node.cpuUsagePercent.toFixed(2)}%</p>
                        <p class="text-gray-600 mb-2"><strong>Memory:</strong> ${node.memoryUsageGB.toFixed(2)} GB / ${node.totalMemoryGB.toFixed(2)} GB</p>
                        <p class="text-gray-600 mb-2"><strong>Disk:</strong> ${node.diskUsageGB.toFixed(2)} GB / ${node.totalDiskGB.toFixed(2)} GB</p>
                        <p class="text-gray-600 mb-2"><strong>Node Status:</strong> ${node.status}</p>
                    </div>
                    <div class="text-sm text-gray-500 mt-4">
                        Last Seen: ${lastSeen}
                    </div>
                </div>
            `;
            nodesContainer.innerHTML += nodeCard;
        });
    };

    // Fetch nodes initially and then every 5 seconds
    fetchNodes();
    setInterval(fetchNodes, 5000); // Poll every 5 seconds
});
