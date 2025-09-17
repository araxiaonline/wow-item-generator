// WoW Theme JavaScript utilities
class WowTheme {
    static init() {
        this.setupTabNavigation();
        this.setupTooltips();
        this.initializeWowheadPower();
    }

    static setupTabNavigation() {
        const tabButtons = document.querySelectorAll('.wow-tab');
        const tabPanels = document.querySelectorAll('.wow-tab-panel');

        tabButtons.forEach(button => {
            button.addEventListener('click', () => {
                const isEntries = button.id === 'tabEntries';

                // Update tab states
                tabButtons.forEach(btn => btn.classList.remove('wow-tab-active'));
                button.classList.add('wow-tab-active');

                // Show/hide panels
                document.getElementById('entriesPanel').classList.toggle('hidden', !isEntries);
                document.getElementById('searchPanel').classList.toggle('hidden', isEntries);
            });
        });
    }

    static setupTooltips() {
        // Initialize Wowhead tooltips when they become available
        if (window.$WowheadPower) {
            window.$WowheadPower.init();
        }
    }

    static initializeWowheadPower() {
        // Configure Wowhead Power settings
        if (window.$WH) {
            window.$WH.Tooltip.config = {
                colorLinks: true,
                iconizeLinks: true,
                renameLinks: true
            };
        }
    }

    static showLoading() {
        document.getElementById('loadingSpinner').classList.remove('hidden');
        document.getElementById('resultsPanel').classList.add('hidden');
    }

    static hideLoading() {
        document.getElementById('loadingSpinner').classList.add('hidden');
    }

    static showResults() {
        document.getElementById('resultsPanel').classList.remove('hidden');
        this.hideLoading();
    }

    static clearResults() {
        const container = document.getElementById('resultsContainer');
        container.innerHTML = '';
        document.getElementById('resultsPanel').classList.add('hidden');
    }

    static getQualityClass(quality) {
        const qualityMap = {
            0: 'quality-poor',
            1: 'quality-common',
            2: 'quality-uncommon',
            3: 'quality-rare',
            4: 'quality-epic',
            5: 'quality-legendary'
        };
        return qualityMap[quality] || 'quality-common';
    }

    static getQualityName(quality) {
        const qualityNames = {
            0: 'Poor',
            1: 'Common',
            2: 'Uncommon',
            3: 'Rare',
            4: 'Epic',
            5: 'Legendary'
        };
        return qualityNames[quality] || 'Common';
    }

    static createItemCard(original, scaled, referenceItem = null) {
        const qualityClass = this.getQualityClass(scaled.quality);
        const cardClass = scaled.quality === 4 ? 'wow-item-epic' : 'wow-item-legendary';

        return `
            <div class="wow-item-card ${cardClass}">
                <div class="wow-item-comparison">
                    <div class="wow-item-before">
                        <h3 class="wow-item-name ${this.getQualityClass(original.quality)}">
                            ${original.name}
                        </h3>
                        <div class="wow-item-image">
                            <img src="https://wow.zamimg.com/images/wow/icons/medium/${this.getItemIcon(original.entry)}.jpg"
                                 alt="${original.name}"
                                 onerror="this.src='https://wow.zamimg.com/images/wow/icons/medium/inv_misc_questionmark.jpg'">
                        </div>
                        <div class="wow-item-stats">
                            <div class="wow-stat-line">Item Level: ${original.itemLevel || 'Unknown'}</div>
                            <div class="wow-stat-line">Quality: ${this.getQualityName(original.quality)}</div>
                            ${this.formatStats(original.stats)}
                        </div>
                    </div>

                    <div class="wow-item-arrow">➤</div>

                    <div class="wow-item-after">
                        <h3 class="wow-item-name ${qualityClass}">
                            ${scaled.name}
                        </h3>
                        <div class="wow-item-image">
                            <img src="https://wow.zamimg.com/images/wow/icons/medium/${this.getItemIcon(scaled.entry)}.jpg"
                                 alt="${scaled.name}"
                                 onerror="this.src='https://wow.zamimg.com/images/wow/icons/medium/inv_misc_questionmark.jpg'">
                        </div>
                        <div class="wow-item-stats">
                            <div class="wow-stat-line">Item Level: ${scaled.itemLevel}</div>
                            <div class="wow-stat-line">Quality: ${this.getQualityName(scaled.quality)}</div>
                            ${this.formatStats(scaled.stats, original.stats)}
                        </div>
                    </div>
                </div>

                ${referenceItem ? `
                    <div class="mt-4 pt-4 border-t border-gray-600">
                        <p class="text-sm text-gray-400 mb-2">Reference Item Used:</p>
                        <p class="text-yellow-400">${referenceItem.name}</p>
                    </div>
                ` : ''}

                <div class="mt-4 pt-4 border-t border-gray-600 flex justify-between">
                    <span class="text-sm text-gray-400">Entry ID: ${scaled.entry}</span>
                    <button class="wow-button-small wow-button-secondary" onclick="WowTheme.copyToClipboard('${scaled.entry}')">
                        Copy Entry ID
                    </button>
                </div>
            </div>
        `;
    }

    static formatStats(stats, originalStats = null) {
        if (!stats || Object.keys(stats).length === 0) {
            return '<div class="wow-stat-line text-gray-500">No stats available</div>';
        }

        return Object.entries(stats).map(([statName, value]) => {
            let className = 'wow-stat-line';
            let prefix = '';

            if (originalStats && originalStats[statName] !== undefined) {
                const originalValue = originalStats[statName];
                if (value > originalValue) {
                    className += ' wow-stat-improved';
                    prefix = '+';
                } else if (value < originalValue) {
                    className += ' wow-stat-decreased';
                }
            }

            return `<div class="${className}">${prefix}${value} ${statName}</div>`;
        }).join('');
    }

    static getItemIcon(entryId) {
        // This would ideally come from your backend or a lookup table
        // For now, return a default or use the entry ID as a fallback
        return `inv_misc_questionmark`;
    }

    static async copyToClipboard(text) {
        try {
            await navigator.clipboard.writeText(text);
            this.showNotification('Copied to clipboard!', 'success');
        } catch (err) {
            console.error('Failed to copy: ', err);
            this.showNotification('Failed to copy', 'error');
        }
    }

    static showNotification(message, type = 'info') {
        const notification = document.createElement('div');
        notification.className = `fixed top-4 right-4 p-4 rounded-lg shadow-lg z-50 transition-all duration-300 ${
            type === 'success' ? 'bg-green-600' :
            type === 'error' ? 'bg-red-600' :
            'bg-blue-600'
        } text-white`;
        notification.textContent = message;

        document.body.appendChild(notification);

        // Animate in
        setTimeout(() => notification.classList.add('opacity-100'), 10);

        // Remove after 3 seconds
        setTimeout(() => {
            notification.classList.add('opacity-0');
            setTimeout(() => notification.remove(), 300);
        }, 3000);
    }

    static createSearchResultItem(item) {
        return `
            <div class="wow-search-item" data-entry="${item.entry}" onclick="WowTheme.selectSearchItem(${item.entry}, '${item.name.replace(/'/g, "\\'")}')">
                <div class="wow-item-image" style="width: 32px; height: 32px;">
                    <img src="https://wow.zamimg.com/images/wow/icons/small/${this.getItemIcon(item.entry)}.jpg"
                         alt="${item.name}"
                         style="width: 100%; height: 100%;"
                         onerror="this.src='https://wow.zamimg.com/images/wow/icons/small/inv_misc_questionmark.jpg'">
                </div>
                <div>
                    <div class="font-medium ${this.getQualityClass(item.quality)}">${item.name}</div>
                    <div class="text-xs text-gray-400">Entry ID: ${item.entry}</div>
                </div>
            </div>
        `;
    }

    static selectSearchItem(entryId, itemName) {
        const entriesInput = document.getElementById('entryIds');
        const currentEntries = entriesInput.value.trim();

        if (currentEntries) {
            entriesInput.value = `${currentEntries}, ${entryId}`;
        } else {
            entriesInput.value = entryId.toString();
        }

        // Clear search
        document.getElementById('itemSearch').value = '';
        document.getElementById('searchResults').classList.add('hidden');

        // Switch to entries tab
        document.getElementById('tabEntries').click();

        this.showNotification(`Added ${itemName} to selection`, 'success');
    }
}

// Initialize theme when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
    WowTheme.init();
});