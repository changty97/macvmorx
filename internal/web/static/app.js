// document.addEventListener('DOMContentLoaded', () => {
//     const nodesContainer = document.getElementById('nodes-container');
//     const loadingIndicator = document.getElementById('loading');
//     const errorDisplay = document.getElementById('error');

//     const fetchNodes = async () => {
//         nodesContainer.innerHTML = ''; // Clear previous nodes
//         loadingIndicator.classList.remove('hidden');
//         errorDisplay.classList.add('hidden');

//         try {
//             const response = await fetch('/api/nodes');
//             if (!response.ok) {
//                 throw new Error(`HTTP error! status: ${response.status}`);
//             }
//             const nodes = await response.json();
//             displayNodes(nodes);
//         } catch (error) {
//             console.error('Failed to fetch nodes:', error);
//             errorDisplay.classList.remove('hidden');
//         } finally {
//             loadingIndicator.classList.add('hidden');
//         }
//     };

//     const displayNodes = (nodes) => {
//         if (nodes.length === 0) {
//             nodesContainer.innerHTML = '<p class="text-center text-gray-600 col-span-full">No Mac Mini nodes registered yet. Waiting for heartbeats...</p>';
//             return;
//         }

//         nodes.sort((a, b) => a.nodeId.localeCompare(b.nodeId)); // Sort by NodeID

//         nodes.forEach(node => {
//             const statusColor = node.isOnline ? 'bg-green-500' : 'bg-red-500';
//             const statusText = node.isOnline ? 'Online' : 'Offline';
//             const lastSeen = new Date(node.lastSeen).toLocaleString();

//             const nodeCard = `
//                 <div class="bg-white rounded-lg shadow-lg p-6 flex flex-col justify-between">
//                     <div>
//                         <h2 class="text-2xl font-semibold text-gray-900 mb-2">${node.nodeId}</h2>
//                         <div class="flex items-center mb-4">
//                             <span class="w-3 h-3 rounded-full ${statusColor} mr-2"></span>
//                             <span class="text-lg font-medium text-gray-700">${statusText}</span>
//                         </div>
//                         <p class="text-gray-600 mb-2"><strong>VMs Running:</strong> ${node.vmCount}/2</p>
//                         <p class="text-gray-600 mb-2"><strong>CPU Usage:</strong> ${node.cpuUsagePercent.toFixed(2)}%</p>
//                         <p class="text-gray-600 mb-2"><strong>Memory:</strong> ${node.memoryUsageGB.toFixed(2)} GB / ${node.totalMemoryGB.toFixed(2)} GB</p>
//                         <p class="text-gray-600 mb-2"><strong>Disk:</strong> ${node.diskUsageGB.toFixed(2)} GB / ${node.totalDiskGB.toFixed(2)} GB</p>
//                         <p class="text-gray-600 mb-2"><strong>Node Status:</strong> ${node.status}</p>
//                     </div>
//                     <div class="text-sm text-gray-500 mt-4">
//                         Last Seen: ${lastSeen}
//                     </div>
//                 </div>
//             `;
//             nodesContainer.innerHTML += nodeCard;
//         });
//     };

//     // Fetch nodes initially and then every 5 seconds
//     fetchNodes();
//     setInterval(fetchNodes, 5000); // Poll every 5 seconds
// });

const nodesTabBtn = document.getElementById('nodesTabBtn');
        const jobsTabBtn = document.getElementById('jobsTabBtn');
        const nodesTabContent = document.getElementById('nodesTabContent');
        const jobsTabContent = document.getElementById('jobsTabContent');
        const nodesGrid = document.getElementById('nodesGrid');
        const jobsTableBody = document.getElementById('jobsTableBody');
        const refreshIntervalSelect = document.getElementById('refreshInterval');

        const ORCHESTRATOR_BASE_URL = 'http://localhost:8080';

        let activeTab = 'nodes';
        let refreshTimer;

        function showTab(tabName) {
            nodesTabBtn.classList.remove('active');
            jobsTabBtn.classList.remove('active');
            nodesTabContent.classList.add('hidden');
            jobsTabContent.classList.add('hidden');

            clearInterval(refreshTimer);

            if (tabName === 'nodes') {
                nodesTabBtn.classList.add('active');
                nodesTabContent.classList.remove('hidden');
                activeTab = 'nodes';
                fetchNodes();
            } else if (tabName === 'jobs') {
                jobsTabBtn.classList.add('active');
                jobsTabContent.classList.remove('hidden');
                activeTab = 'jobs';
                fetchJobs();
            }
            setRefreshInterval();
        }

        nodesTabBtn.addEventListener('click', () => showTab('nodes'));
        jobsTabBtn.addEventListener('click', () => showTab('jobs'));

        function setRefreshInterval() {
            const interval = parseInt(refreshIntervalSelect.value, 10);
            clearInterval(refreshTimer);
            refreshTimer = setInterval(() => {
                if (activeTab === 'nodes') {
                    fetchNodes();
                } else if (activeTab === 'jobs') {
                    fetchJobs();
                }
            }, interval);
            console.log(`Refresh interval set to ${interval / 1000} seconds.`);
        }

        refreshIntervalSelect.addEventListener('change', setRefreshInterval);

        function formatDuration(seconds) {
            if (seconds === undefined || seconds === null) return 'N/A';
            const s = Math.floor(seconds % 60);
            const m = Math.floor((seconds / 60) % 60);
            const h = Math.floor(seconds / 3600);
            if (h > 0) return `${h}h ${m}m ${s}s`;
            if (m > 0) return `${m}m ${s}s`;
            return `${s}s`;
        }

        function getStatusClass(status) {
            switch (status) {
                case 'online':
                case 'healthy':
                case 'completed':
                    return 'status-online';
                case 'offline':
                case 'failed':
                case 'cancelled':
                case 'skipped':
                    return 'status-offline';
                case 'provisioning':
                    return 'status-provisioning';
                case 'running':
                    return 'status-running';
                case 'queued':
                    return 'status-queued';
                default:
                    return 'status-unknown';
            }
        }

        async function fetchNodes() {
            nodesGrid.innerHTML = '<p class="text-gray-500">Loading node data...</p>';
            try {
                const url = `${ORCHESTRATOR_BASE_URL}/api/nodes`;
                console.log('Fetching nodes from:', url);
                const response = await fetch(url);
                
                if (!response.ok) {
                    const errorText = await response.text();
                    throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`);
                }

                const nodes = await response.json();
                nodesGrid.innerHTML = '';

                if (nodes.length === 0) {
                    nodesGrid.innerHTML = '<p class="text-gray-500">No nodes registered yet.</p>';
                    return;
                }

                nodes.forEach(node => {
                    const lastSeen = new Date(node.LastSeen);
                    const timeSinceLastSeen = node.IsOnline ? 'Just now' : formatDuration((Date.now() - lastSeen.getTime()) / 1000) + ' ago';
                    const statusClass = getStatusClass(node.IsOnline ? 'online' : 'offline');

                    // FIX: Ensure node.VMs is an array, even if undefined or null
                    const vms = node.VMs || [];
                    const vmList = vms.length > 0
                        ? vms.map(vm => `<li class="text-sm"><strong>VM ID:</strong> ${vm.VMID} (${vm.ImageName}) - ${vm.VMIPAddress} - ${formatDuration(vm.RuntimeSeconds)}</li>`).join('')
                        : '<li class="text-sm text-gray-500">No VMs running.</li>';

                    nodesGrid.innerHTML += `
                        <div class="card p-6 flex flex-col space-y-4">
                            <h3 class="text-lg font-semibold text-gray-800">${node.nodeId}</h3>
                            <div class="flex items-center space-x-2">
                                <span class="text-sm font-medium ${statusClass}">${node.IsOnline ? 'Online' : 'Offline'}</span>
                                <span class="text-xs text-gray-500">(${timeSinceLastSeen})</span>
                            </div>
                            <div class="grid grid-cols-2 gap-2 text-sm text-gray-700">
                                <p><strong>VMs:</strong> ${node.vmCount}</p>
                                <p><strong>CPU:</strong> ${node.cpuUsagePercent.toFixed(1)}%</p>
                                <p><strong>Memory:</strong> ${node.memoryUsageGB.toFixed(1)} GB / ${node.totalMemoryGB.toFixed(1)} GB</p>
                                <p><strong>Disk:</strong> ${node.diskUsageGB.toFixed(1)} GB / ${node.diskUsageGB.toFixed(1)} GB</p>
                            </div>
                            <div>
                                <p class="text-sm font-medium text-gray-700 mb-1">Running VMs:</p>
                                <ul class="list-disc list-inside text-gray-600">
                                    ${vmList}
                                </ul>
                            </div>
                            <div>
                                <p class="text-sm font-medium text-gray-700 mb-1">Cached Images:</p>
                                <ul class="list-disc list-inside text-gray-600">
                                    ${node.cachedImages.length > 0 ? node.cachedImages.map(img => `<li class="text-sm">${img}</li>`).join('') : '<li class="text-sm text-gray-500">No images cached.</li>'}
                                </ul>
                            </div>
                        </div>
                    `;
                });
            } catch (error) {
                console.error('Error fetching nodes:', error);
                nodesGrid.innerHTML = '<p class="text-red-500">Failed to load node data. Please ensure the orchestrator is running and accessible at ' + ORCHESTRATOR_BASE_URL + '</p>';
            }
        }

        async function fetchJobs() {
            jobsTableBody.innerHTML = '<tr><td colspan="8" class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">Loading job data...</td></tr>';
            try {
                const url = `${ORCHESTRATOR_BASE_URL}/api/jobs`;
                console.log('Fetching jobs from:', url);
                const response = await fetch(url);

                if (!response.ok) {
                    const errorText = await response.text();
                    throw new Error(`HTTP error! status: ${response.status}, message: ${errorText}`);
                }

                const jobs = await response.json();
                jobsTableBody.innerHTML = '';

                if (jobs.length === 0) {
                    jobsTableBody.innerHTML = '<tr><td colspan="8" class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">No GitHub jobs tracked yet.</td></tr>';
                    return;
                }

                jobs.sort((a, b) => new Date(b.QueueTime) - new Date(a.QueueTime));

                jobs.forEach(job => {
                    let runtime = 'N/A';
                    if (job.VMStartTime) {
                        const start = new Date(job.VMStartTime);
                        const end = job.EndTime ? new Date(job.EndTime) : new Date();
                        runtime = formatDuration((end.getTime() - start.getTime()) / 1000);
                    } else if (job.ProvisioningStartTime) {
                         const start = new Date(job.ProvisioningStartTime);
                         const end = job.EndTime ? new Date(job.EndTime) : new Date();
                         runtime = formatDuration((end.getTime() - start.getTime()) / 1000);
                    }

                    const statusClass = getStatusClass(job.Status);
                    const labels = job.Labels && job.Labels.length > 0 ? job.Labels.join(', ') : 'None';

                    jobsTableBody.innerHTML += `
                        <tr class="hover:bg-gray-50">
                            <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900">${job.JobID}</td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-900">${job.RunnerName || 'N/A'}</td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${job.ImageName}</td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm font-semibold ${statusClass}">${job.Status.charAt(0).toUpperCase() + job.Status.slice(1)}</td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${job.NodeID || 'Pending'}</td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${job.VMIPAddress || 'Pending'}</td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${runtime}</td>
                            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">${labels}</td>
                        </tr>
                    `;
                });
            } catch (error) {
                console.error('Error fetching jobs:', error);
                jobsTableBody.innerHTML = '<tr><td colspan="8" class="px-6 py-4 whitespace-nowrap text-sm text-red-500">Failed to load job data. Please ensure the orchestrator is running and accessible at ' + ORCHESTRATOR_BASE_URL + '</td></tr>';
            }
        }

        // Initial load based on active tab
        showTab(activeTab);

        // Set initial refresh interval based on default selected value
        setRefreshInterval();
