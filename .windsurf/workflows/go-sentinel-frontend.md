---
description: Go Sentinel Frontend Template Development Workflow
---

# Go Sentinel Frontend Template Development Workflow

This workflow defines the systematic approach to frontend template development in the Go Sentinel project, combining TDD principles with Go's HTML templating best practices.

## Template Structure and Inheritance Model

### 1. Three-Tier Template Hierarchy (✅ Must Follow)

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

The base template (`layouts/base.tmpl`) defines these standard blocks:
- `{{block "title" .}}` - For page title
- `{{block "head" .}}` - For page-specific head elements
- `{{block "sidebar" .}}` - For sidebar customization
- `{{block "header-extras" .}}` - For header additional content
- `{{block "content" .}}` - For main page content
- `{{block "scripts" .}}` - For page-specific scripts

### 3. Page Template Implementation

Every page template must:
1. Define itself with: `{{define "pages/pagename"}}{{template "base" .}}{{end}}`
2. Implement required blocks like: `{{define "content"}}...{{end}}`
3. Not use dynamic template name construction (use explicit template names)
4. Include all necessary blocks even if empty

## Development Workflow

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
   - Add minimal required content to pass tests

3. **Iterate with Additional Tests**
   - Test responsive behavior with different viewports
   - Test dynamic content conditions
   - Test accessibility compliance

### Phase 2: Template Composition and Integration

1. **Extract Common Patterns to Partials**
   - Identify repeating UI patterns
   - Create partial templates
   - Reference from page templates: `{{template "partials/component" .}}`

2. **Add Layout Variations as Needed**
   - Create alternative layouts in `layouts/` directory
   - Implement proper inheritance

### Phase 3: Server Integration

1. **Template Loading Order**
   Always maintain strict loading order:
   ```go
   // Parse in order: layouts → partials → pages
   layouts, _ := filepath.Glob(filepath.Join(templatePath, "layouts", "*.tmpl"))
   partials, _ := filepath.Glob(filepath.Join(templatePath, "partials", "*.tmpl"))
   pages, _ := filepath.Glob(filepath.Join(templatePath, "pages", "*.tmpl"))
   
   tmpl = template.New("").Funcs(funcMap)
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

## Troubleshooting Common Issues

### Template Visibility Problems

**Issue:** Content not appearing on pages
**Debug Steps:**
1. Verify correct template loading order
2. Check template naming conventions
3. Ensure blocks are properly defined in page templates
4. Verify data being passed to templates

### Template Execution Errors

**Issue:** Template syntax or execution errors
**Debug Steps:**
1. Check for mismatched define/end tags
2. Verify variable usage with proper dot notation
3. Ensure templates referenced with {{template}} exist
4. Check for nil data when required

## Response Structure for Template Tasks

When completing template-related tasks, follow this structured format:

### 1. Summary Header
Brief statement of what was accomplished, including which templates were created or modified.

### 2. Completed Work (✅)
Detailed list of completed items with checkmarks:
- **Tests Implemented**: Template tests written
- **Templates Created/Modified**: Which template files were changed and how
- **Server Integration**: How templates are loaded and executed
- **Roadmap Updates**: Which roadmap items were marked complete

### 3. Technical Details
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
