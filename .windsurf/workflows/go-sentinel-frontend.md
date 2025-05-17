---
description: Go Sentinel Frontend Template Development Workflow
includes: high-confidence-coding.md
---

# Go Sentinel Frontend Template Development Workflow

This workflow combines TDD principles with Go's HTML templating best practices.

## Template Structure and Inheritance Model

### 1. Three-Tier Template Hierarchy

```
templates/
├── layouts/      # Base templates with overall structure
│   └── base.tmpl # The master template with block definitions
├── pages/        # Page-specific templates 
│   └── *.tmpl    # One template per route/page
└── partials/     # Reusable UI components
    └── *.tmpl    # Small, focused component templates
```

### 2. Base Template Block Structure

The base template defines standard blocks:
- `{{block "title" .}}` - For page title
- `{{block "head" .}}` - For page-specific head elements
- `{{block "content" .}}` - For main page content
- `{{block "scripts" .}}` - For page-specific scripts

## Development Workflow

> **All phases below must follow the High-Confidence Coding Workflow.**

### Phase 1: Test-First Template Development

1. **Write Template Tests First**
   ```go
   func TestDashboardTemplate(t *testing.T) {
       // Given a dashboard template with test data
       data := map[string]interface{}{
           "Title": "Dashboard",
           "ActivePage": "dashboard",
       }
       
       // When we render the template
       html, err := renderTemplate("pages/dashboard", data)
       
       // Then the rendered output should:
       require.NoError(t, err)
       assert.Contains(t, html, "<title>Dashboard")
       assert.Contains(t, html, `class="dashboard-container"`)
   }
   ```

2. **Implement Template Structure**
   - Create page template with required blocks
   - Define inheritance from base template

### Phase 2: Template Composition and Integration

1. **Extract Common Patterns to Partials**
   - Identify repeating UI patterns
   - Create partial templates

2. **Add Layout Variations as Needed**
   - Create alternative layouts in `layouts/` directory

### Phase 3: Server Integration

1. **Template Loading Order**
   Always maintain strict loading order:
   ```go
   // Parse in order: layouts → partials → pages
   tmpl = template.Must(tmpl.ParseFiles(layouts...))
   tmpl = template.Must(tmpl.ParseFiles(partials...))
   tmpl = template.Must(tmpl.ParseFiles(pages...))
   ```

2. **Use Named Template Execution**
   ```go
   func (s *Server) render(w http.ResponseWriter, name string, data interface{}) {
       // Always execute the page template, not the blocks directly
       err := s.templates.ExecuteTemplate(w, name, data)
       if err != nil {
           http.Error(w, err.Error(), http.StatusInternalServerError)
       }
   }
   ```

## Response Structure for Template Tasks

When completing template-related tasks, follow this structured format:

### 1. Summary Header
Brief statement of what was accomplished.

### 2. Completed Work (✅)
- **Tests Implemented**: Template tests written
- **Templates Created/Modified**: Which template files were changed
- **Roadmap Updates**: Which roadmap items were marked complete

### 3. Technical Details

### 4. High-Confidence Checkpoint
- Complete the confidence checklist as per [high-confidence-coding.md]
- Ensure ≥95% test coverage, all validations pass, and no speculative code remains
- Document reasoning, edge cases, and any uncertainties
- If confidence is <95%, halt and request clarification or peer review before merging

- Template inheritance patterns used
- Reusable components created
- Accessibility considerations
- Mobile responsiveness

### 4. Next Steps
- Clear prioritized list of upcoming template tasks
- Any potential template challenges to address
- Suggested approach for the next implementation

## Best Practices & Common Pitfalls

### ✅ Do These:
- Write tests first, then template code
- Keep templates DRY using partials
- Use explicit template names, not dynamic construction
- Include default content in blocks where appropriate
- Follow proper template inheritance patterns

### ❌ Avoid These:
- Dynamically constructing template names
- Using {{with}} blocks with potential nil values
- Deeply nesting templates without proper data passing
- Duplicating code across multiple templates
- Mixing display logic and business logic in templates

Do not break my server, the render is working as expected, don't modify it, adapt to my configurations and work with it. 