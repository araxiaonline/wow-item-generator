// Main Item Scaler Application Logic
class ItemScaler {
    constructor() {
        this.backendUrl = 'http://localhost:8080'; // Update this to match your backend
        this.searchTimeout = null;
        this.init();
    }

    init() {
        this.setupEventListeners();
        this.loadDefaultValues();
    }

    setupEventListeners() {
        // Scale button
        document.getElementById('scaleButton').addEventListener('click', () => {
            this.scaleItems();
        });

        // Clear button
        document.getElementById('clearButton').addEventListener('click', () => {
            WowTheme.clearResults();
        });

        // Export SQL button
        document.getElementById('exportSqlButton').addEventListener('click', () => {
            this.exportSQL();
        });

        // Copy results button
        document.getElementById('copyResultsButton').addEventListener('click', () => {
            this.copyResults();
        });

        // Search input
        document.getElementById('itemSearch').addEventListener('input', (e) => {
            this.handleSearchInput(e.target.value);
        });

        // Hide search results when clicking outside
        document.addEventListener('click', (e) => {
            const searchContainer = document.getElementById('searchPanel');
            if (!searchContainer.contains(e.target)) {
                document.getElementById('searchResults').classList.add('hidden');
            }
        });
    }

    loadDefaultValues() {
        // Load any saved preferences from localStorage
        const savedValues = this.getSavedValues();
        if (savedValues) {
            document.getElementById('itemLevel').value = savedValues.itemLevel || 325;
            document.getElementById('phase').value = savedValues.phase || 0;
            document.getElementById('quality').value = savedValues.quality || 4;
            document.getElementById('catchup').value = savedValues.catchup || 1.0;
        }
    }

    getSavedValues() {
        try {
            return JSON.parse(localStorage.getItem('wowItemScalerSettings'));
        } catch {
            return null;
        }
    }

    saveValues() {
        const values = {
            itemLevel: parseInt(document.getElementById('itemLevel').value),
            phase: parseInt(document.getElementById('phase').value),
            quality: parseInt(document.getElementById('quality').value),
            catchup: parseFloat(document.getElementById('catchup').value)
        };
        localStorage.setItem('wowItemScalerSettings', JSON.stringify(values));
    }

    async scaleItems() {
        try {
            WowTheme.showLoading();
            this.saveValues();

            const scalingParams = this.getScalingParameters();
            const itemEntries = this.getItemEntries();

            if (itemEntries.length === 0) {
                WowTheme.hideLoading();
                WowTheme.showNotification('Please provide item entries or search for items', 'error');
                return;
            }

            const results = await this.callBackendAPI(scalingParams, itemEntries);
            this.displayResults(results);

        } catch (error) {
            console.error('Error scaling items:', error);
            WowTheme.hideLoading();
            WowTheme.showNotification(`Error: ${error.message}`, 'error');
        }
    }

    getScalingParameters() {
        return {
            itemLevel: parseInt(document.getElementById('itemLevel').value),
            phase: parseInt(document.getElementById('phase').value),
            quality: parseInt(document.getElementById('quality').value),
            catchup: parseFloat(document.getElementById('catchup').value)
        };
    }

    getItemEntries() {
        const entriesText = document.getElementById('entryIds').value.trim();
        if (!entriesText) {
            return [];
        }

        return entriesText
            .split(',')
            .map(entry => entry.trim())
            .filter(entry => entry && !isNaN(entry))
            .map(entry => parseInt(entry));
    }

    async callBackendAPI(scalingParams, itemEntries) {
        const requestData = {
            ...scalingParams,
            entries: itemEntries
        };

        console.log('Sending request:', requestData);

        const response = await fetch(`${this.backendUrl}/api/scale-items`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(requestData)
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.message || `HTTP ${response.status}: ${response.statusText}`);
        }

        return await response.json();
    }

    displayResults(results) {
        const container = document.getElementById('resultsContainer');
        container.innerHTML = '';

        if (!results || !results.items || results.items.length === 0) {
            container.innerHTML = `
                <div class="text-center text-gray-400 py-8">
                    <p>No items were successfully scaled.</p>
                    <p class="text-sm mt-2">Check the console for detailed error information.</p>
                </div>
            `;
            WowTheme.showResults();
            return;
        }

        // Display summary
        const summary = `
            <div class="wow-panel mb-6">
                <div class="wow-panel-content">
                    <div class="text-center">
                        <h3 class="text-xl font-semibold text-yellow-400 mb-4">Scaling Summary</h3>
                        <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
                            <div>
                                <div class="text-gray-400">Total Items</div>
                                <div class="text-white font-semibold">${results.summary?.total || results.items.length}</div>
                            </div>
                            <div>
                                <div class="text-gray-400">Successful</div>
                                <div class="text-green-400 font-semibold">${results.summary?.successful || results.items.length}</div>
                            </div>
                            <div>
                                <div class="text-gray-400">Failed</div>
                                <div class="text-red-400 font-semibold">${results.summary?.failed || 0}</div>
                            </div>
                            <div>
                                <div class="text-gray-400">Success Rate</div>
                                <div class="text-yellow-400 font-semibold">${results.summary?.successRate || '100.0'}%</div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        `;

        container.innerHTML = summary;

        // Display individual item results
        results.items.forEach(result => {
            if (result.success) {
                const itemCard = WowTheme.createItemCard(
                    result.original,
                    result.scaled,
                    result.referenceItem
                );
                container.innerHTML += itemCard;
            } else {
                // Display error for failed item
                container.innerHTML += `
                    <div class="wow-item-card border-red-600">
                        <div class="text-center">
                            <h3 class="text-red-400 font-semibold mb-2">Failed to scale item</h3>
                            <p class="text-gray-300 mb-2">Entry ID: ${result.entryId}</p>
                            <div class="text-sm text-gray-400">
                                ${result.errors ? result.errors.map(error => `<p>${error}</p>`).join('') : 'Unknown error occurred'}
                            </div>
                        </div>
                    </div>
                `;
            }
        });

        WowTheme.showResults();
        this.reinitializeTooltips();
    }

    reinitializeTooltips() {
        // Reinitialize Wowhead tooltips for newly added content
        if (window.$WowheadPower) {
            window.$WowheadPower.refreshLinks();
        }
    }

    async handleSearchInput(query) {
        clearTimeout(this.searchTimeout);

        if (!query || query.length < 2) {
            document.getElementById('searchResults').classList.add('hidden');
            return;
        }

        this.searchTimeout = setTimeout(async () => {
            try {
                const results = await this.searchItems(query);
                this.displaySearchResults(results);
            } catch (error) {
                console.error('Search error:', error);
                WowTheme.showNotification('Search failed', 'error');
            }
        }, 300);
    }

    async searchItems(query) {
        const response = await fetch(`${this.backendUrl}/api/search-items?q=${encodeURIComponent(query)}`, {
            method: 'GET',
            headers: {
                'Content-Type': 'application/json',
            }
        });

        if (!response.ok) {
            throw new Error(`Search failed: ${response.status}`);
        }

        return await response.json();
    }

    displaySearchResults(results) {
        const container = document.getElementById('searchResults');

        if (!results || results.length === 0) {
            container.innerHTML = `
                <div class="wow-search-item">
                    <div class="text-gray-400">No items found</div>
                </div>
            `;
        } else {
            container.innerHTML = results
                .slice(0, 10) // Limit to 10 results
                .map(item => WowTheme.createSearchResultItem(item))
                .join('');
        }

        container.classList.remove('hidden');
    }

    async exportSQL() {
        try {
            const scalingParams = this.getScalingParameters();
            const itemEntries = this.getItemEntries();

            if (itemEntries.length === 0) {
                WowTheme.showNotification('No items to export', 'error');
                return;
            }

            const response = await fetch(`${this.backendUrl}/api/export-sql`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    ...scalingParams,
                    entries: itemEntries
                })
            });

            if (!response.ok) {
                throw new Error(`Export failed: ${response.status}`);
            }

            const blob = await response.blob();
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            a.download = `scaled_items_${Date.now()}.sql`;
            document.body.appendChild(a);
            a.click();
            document.body.removeChild(a);
            window.URL.revokeObjectURL(url);

            WowTheme.showNotification('SQL exported successfully!', 'success');

        } catch (error) {
            console.error('Export error:', error);
            WowTheme.showNotification('Export failed', 'error');
        }
    }

    async copyResults() {
        try {
            const container = document.getElementById('resultsContainer');
            const textContent = container.innerText;
            await navigator.clipboard.writeText(textContent);
            WowTheme.showNotification('Results copied to clipboard!', 'success');
        } catch (error) {
            console.error('Copy error:', error);
            WowTheme.showNotification('Failed to copy results', 'error');
        }
    }
}

// Initialize the application
document.addEventListener('DOMContentLoaded', () => {
    new ItemScaler();
});