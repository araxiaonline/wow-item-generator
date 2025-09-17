# WoW Item Scaler Web Frontend

A beautiful World of Warcraft-themed web interface for scaling items using your existing Go backend.

## Features

- **Immersive WoW-styled UI** with layered backgrounds, magical particle effects, and authentic textures
- **Modern visual effects** including animated glowing elements, parallax backgrounds, and fantasy decorations
- **Multiple input methods**:
  - Single item by entry ID
  - Comma-separated list of entry IDs
  - Search by item name with dropdown selection
- **Visual item comparison** showing before/after scaling
- **Wowhead integration** for item images and tooltips
- **Responsive design** that works on desktop and mobile
- **Export functionality** for SQL and copy-to-clipboard

## Dependencies & Libraries Used

### External CDN Libraries
1. **Tailwind CSS** (`https://cdn.tailwindcss.com`)
   - Utility-first CSS framework
   - Used for responsive layouts, spacing, and utility classes
   - No build process required

2. **Wowhead Power** (`https://wow.zamimg.com/widgets/power.js`)
   - Official Wowhead tooltips and item integration
   - Provides item images and tooltip functionality
   - Authentic WoW item data

3. **Google Fonts**
   - **Cinzel**: Elegant serif font for body text and UI elements
   - **Uncial Antiqua**: Fantasy/medieval font for titles and headings
   - Provides authentic WoW-like typography

### Custom Theme System
- **Custom CSS Variables** for consistent WoW color scheme
- **Modular component classes** (`.wow-panel`, `.wow-button`, etc.)
- **Quality-based item coloring** (epic purple, legendary orange)
- **Animated effects** and hover states

## Setup

1. Open `index.html` in a web browser
2. Update the backend URL in `js/item-scaler.js` (line 4) to match your Go backend server
3. Ensure your backend provides these API endpoints:
   - `POST /api/scale-items` - Scale items with parameters
   - `GET /api/search-items?q=query` - Search items by name
   - `POST /api/export-sql` - Export scaled items as SQL

## Backend API Expected Format

### Scale Items Request
```json
{
  "itemLevel": 325,
  "phase": 1,
  "quality": 4,
  "catchup": 1.5,
  "entries": [19019, 19103, 17076]
}
```

### Scale Items Response
```json
{
  "items": [
    {
      "success": true,
      "entryId": 19019,
      "original": {
        "entry": 19019,
        "name": "Thunderfury, Blessed Blade of the Windseeker",
        "quality": 5,
        "itemLevel": 80,
        "stats": { "Stamina": 15, "Agility": 5 }
      },
      "scaled": {
        "entry": 19019,
        "name": "Thunderfury, Blessed Blade of the Windseeker",
        "quality": 5,
        "itemLevel": 325,
        "stats": { "Stamina": 45, "Agility": 15 }
      },
      "referenceItem": {
        "name": "Some Reference Item",
        "entry": 12345
      }
    }
  ],
  "summary": {
    "total": 1,
    "successful": 1,
    "failed": 0,
    "successRate": "100.0"
  }
}
```

### Search Items Response
```json
[
  {
    "entry": 19019,
    "name": "Thunderfury, Blessed Blade of the Windseeker",
    "quality": 5
  }
]
```

## File Structure

```
wow-item-scaler-web/
├── index.html          # Main HTML file with CDN dependencies
├── css/
│   └── wow-theme.css   # Complete WoW theme styles (8.4KB)
├── js/
│   ├── wow-theme.js    # Theme utilities and UI components (9.8KB)
│   └── item-scaler.js  # Main application logic and API calls (12KB)
└── README.md           # This documentation
```

## Technical Details

### CSS Architecture
- **CSS Custom Properties** for theme consistency
- **Component-based styling** with `.wow-` prefixed classes
- **Layered background system** with multiple gradients, textures, and particle effects
- **Animated decorative elements** including glowing gems, magical sparkles, and pulsing effects
- **Responsive breakpoints** for mobile compatibility
- **Advanced visual effects** using pure CSS (no external images required)

### JavaScript Architecture
- **Class-based ES6 modules** for organization
- **Event-driven architecture** for user interactions
- **LocalStorage integration** for user preferences
- **Error handling** with user-friendly notifications

### Browser Compatibility
- Modern browsers with ES6+ support
- Chrome, Firefox, Safari, Edge
- Mobile browsers (responsive design)

## Customization

### Backend URL
Update the `backendUrl` in `js/item-scaler.js`:
```javascript
this.backendUrl = 'http://your-backend-server:8080';
```

### Styling
The WoW theme can be customized in `css/wow-theme.css`. Key variables:
```css
:root {
    --wow-gold: #ffd100;        /* Primary accent color */
    --wow-blue: #0084ff;        /* Rare item color */
    --wow-purple: #a335ee;      /* Epic item color */
    --wow-orange: #ff8000;      /* Legendary item color */
    --wow-bg-dark: #0a0a0a;     /* Primary background */
    --wow-bg-panel: #1a1a1a;    /* Panel background */
}
```

### Item Icons
Item icons are loaded from Wowhead. Enhance the `getItemIcon()` function in `wow-theme.js` with your own icon mapping logic.

## No Build Process Required
This is a pure frontend application that runs directly in the browser:
- No bundlers (Webpack, Vite, etc.)
- No transpilation needed
- No package.json or npm install
- Just open `index.html` and go!