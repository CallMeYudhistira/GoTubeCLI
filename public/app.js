document.addEventListener('DOMContentLoaded', () => {
    const urlInput = document.getElementById('urlInput');
    const downloadForm = document.getElementById('downloadForm');
    const typeSelect = document.getElementById('typeSelect');
    const downloadBtn = document.getElementById('downloadBtn');
    
    // UI Elements
    const videoInfo = document.getElementById('videoInfo');
    const videoThumb = document.getElementById('videoThumb');
    const videoTitle = document.getElementById('videoTitle');
    const videoDuration = document.getElementById('videoDuration');
    
    const progressContainer = document.getElementById('progressContainer');
    const progressBar = document.getElementById('progressBar');
    const progressPercentage = document.getElementById('progressPercentage');
    const statusLabel = document.getElementById('statusLabel');
    
    const errorBox = document.getElementById('errorBox');
    const completedBox = document.getElementById('completedBox');
    const downloadLink = document.getElementById('downloadLink');

    let debounceTimer;
    let pollInterval;

    // Handle URL paste / input
    urlInput.addEventListener('input', (e) => {
        clearTimeout(debounceTimer);
        const url = e.target.value.trim();
        
        if (!url || !url.startsWith('http')) {
            videoInfo.classList.add('hidden');
            return;
        }

        debounceTimer = setTimeout(() => {
            fetchVideoInfo(url);
        }, 500);
    });

    // Handle form submit
    downloadForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const url = urlInput.value.trim();
        const type = typeSelect.value;

        if (!url) return;

        resetUI();
        setLoading(true);

        try {
            const res = await fetch('/api/download', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ url, type })
            });

            const data = await res.json();

            if (!res.ok) {
                throw new Error(data.error || 'Failed to start download');
            }

            // Start polling
            startPolling(data.id);
            
        } catch (err) {
            showError(err.message);
            setLoading(false);
        }
    });

    async function fetchVideoInfo(url) {
        try {
            const res = await fetch(`/api/info?url=${encodeURIComponent(url)}`);
            const data = await res.json();
            
            if (res.ok) {
                videoTitle.textContent = data.title;
                videoDuration.textContent = data.duration;
                videoThumb.src = data.thumbnail;
                videoInfo.classList.remove('hidden');
                errorBox.classList.add('hidden');
            }
        } catch (err) {
            console.error('Failed to fetch info', err);
        }
    }

    function startPolling(jobId) {
        progressContainer.classList.remove('hidden');
        
        pollInterval = setInterval(async () => {
            try {
                const res = await fetch(`/api/progress?id=${jobId}`);
                const data = await res.json();

                if (!res.ok) throw new Error(data.error || 'Failed to fetch progress');

                // Update Progress UI
                progressBar.style.width = `${data.progress}%`;
                progressPercentage.textContent = `${data.progress}%`;
                
                // Format status beautifully
                const formattedStatus = data.status.charAt(0).toUpperCase() + data.status.slice(1);
                statusLabel.textContent = formattedStatus + '...';

                if (data.status === 'completed') {
                    clearInterval(pollInterval);
                    showSuccess(data.file);
                    setLoading(false);
                } else if (data.status === 'error') {
                    clearInterval(pollInterval);
                    showError(data.error || 'An error occurred during download');
                    setLoading(false);
                }

            } catch (err) {
                clearInterval(pollInterval);
                showError(err.message);
                setLoading(false);
            }
        }, 1000);
    }

    function resetUI() {
        errorBox.classList.add('hidden');
        completedBox.classList.add('hidden');
        progressContainer.classList.add('hidden');
        progressBar.style.width = '0%';
        progressPercentage.textContent = '0%';
        if (pollInterval) clearInterval(pollInterval);
    }

    function setLoading(isLoading) {
        downloadBtn.disabled = isLoading;
        downloadBtn.textContent = isLoading ? 'Processing...' : 'Download';
        urlInput.disabled = isLoading;
        typeSelect.disabled = isLoading;
    }

    function showError(msg) {
        errorBox.textContent = msg;
        errorBox.classList.remove('hidden');
        progressContainer.classList.add('hidden');
    }

    function showSuccess(fileUrl) {
        completedBox.classList.remove('hidden');
        downloadLink.href = fileUrl;
        
        // Auto trigger download
        setTimeout(() => {
            downloadLink.click();
        }, 500);
    }
});
