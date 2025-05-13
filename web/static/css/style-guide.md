# Go Sentinel Style Guide

This document outlines the core design patterns and UI components used throughout the Go Sentinel application to ensure consistency across all pages.

## Color System

All colors are defined as CSS variables in the root:

```css
:root {
  --bg-dark: #161820;      /* Main background */
  --bg-darker: #0f111a;    /* Sidebar background */
  --bg-card: #1e2029;      /* Card background */
  --primary: #3b82f6;      /* Primary actions, links */
  --secondary: #6366f1;    /* Secondary actions */
  --success: #10b981;      /* Success states */
  --error: #ef4444;        /* Error states */
  --warning: #f59e0b;      /* Warning states */
  --text: #f3f4f6;         /* Primary text */
  --text-dim: #9ca3af;     /* Secondary text */
  --border: #282a36;       /* Border color */
}
```

## Layout Structure

### Page Layout

Every page follows this structure:
1. Sidebar navigation (fixed width: 240px)
2. Main content area with:
   - Header section with title and description
   - Dashboard container with content sections

```html
<div class="app-container">
  <nav class="sidebar">
    {{template "partials/sidebar" .}}
  </nav>
  <main class="main-content">
    <header class="dashboard-header">
      <h1>{{.Title}}</h1>
      <p>Description text</p>
    </header>
    <div class="dashboard-container">
      <!-- Content sections go here -->
    </div>
  </main>
</div>
```

### Section Structure

All content sections follow this pattern:
1. Container with `.recent-tests` class (despite the name, used for all section types)
2. Header with `.section-header` containing title and optional actions
3. Content container (usually `.test-table-container` or similar)

```html
<section class="recent-tests">
  <div class="section-header">
    <h2>Section Title</h2>
    <!-- Optional action buttons -->
    <button class="run-all-button">Action</button>
  </div>
  <div class="test-table-container">
    <!-- Section content -->
  </div>
</section>
```

## Spacing System

- **Container gap**: 1.5rem (`gap: 1.5rem` in `.dashboard-container`)
- **Section margins**: 1.5rem bottom margin between major sections
- **Section header padding**: 1.25rem 1.5rem
- **Card padding**: 1.5rem
- **Content padding**: 1rem 1.5rem for table cells and content areas

## Typography

- **Page title**: 1.875rem, font-weight 600
- **Section headers**: 1.25rem, font-weight 600
- **Card titles**: 0.875rem, color var(--text-dim)
- **Body text**: 0.875rem for most content
- **Stat values**: 2.25rem, font-weight 700

## Components

### Cards

Cards use these consistent styles:
- Background: var(--bg-card)
- Border-radius: 0.5rem
- Box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1)
- Optional top border for status indication

### Buttons

#### Primary Action Button
Used for main actions like "Run All Tests":
```html
<button class="run-all-button">Run All Tests</button>
```

#### Modern Button System
For all other buttons, use the `.btn` class with modifiers:
```html
<button class="btn">Default</button>
<button class="btn btn-success">Success</button>
<button class="btn btn-error">Error</button>
<button class="btn btn-warning">Warning</button>
<button class="btn btn-info">Info</button>
```

### Status Indicators

#### Status Badges
Used for showing test status:
```html
<span class="status-badge passed">Passed</span>
<span class="status-badge failed">Failed</span>
```

#### Connection Status
For showing connection state:
```html
<span class="status-indicator connected">Connected</span>
<span class="status-indicator disconnected">Disconnected</span>
```

## Tables

Tables use these consistent styles:
- Full width (100%)
- Header: text-align left, padding 1rem 1.5rem, color var(--text-dim)
- Rows: border-bottom 1px solid var(--border)
- Cell padding: 1rem 1.5rem

## Responsive Behavior

- Mobile breakpoint: 768px
- Stack cards vertically on mobile
- Convert section headers to column layout on mobile
- Ensure tables have horizontal scroll on small screens

## Accessibility Features

- Skip link for keyboard navigation
- High contrast mode support
- Reduced motion preference support
- Proper ARIA attributes
- Focus indicators on interactive elements
