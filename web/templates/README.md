# Go Sentinel Templates

This directory contains all server-side templates for the Go Sentinel web interface, following a strict three-tier hierarchy for maintainability and consistency.

## 📁 Directory Structure

```
web/templates/
├── layouts/          # Base templates
│   ├── base.tmpl    # Main layout
│   └── auth.tmpl    # Authentication layout
├── pages/           # Route-specific templates
│   ├── dashboard.tmpl
│   ├── history.tmpl
│   └── settings.tmpl
└── partials/        # Reusable components
    ├── _header.tmpl
    ├── _footer.tmpl
    ├── _test_card.tmpl
    └── _toast.tmpl
```

## 🏗 Template Hierarchy

### 1. Layouts
Base templates that define the overall page structure.

**base.tmpl**
```html
<!DOCTYPE html>
<html>
<head>
    <title>{{template "title" .}} - Go Sentinel</title>
    {{template "head" .}}
</head>
<body>
    {{template "header" .}}
    <main>
        {{template "content" .}}
    </main>
    {{template "footer" .}}
    {{template "scripts" .}}
</body>
</html>
```

### 2. Pages
Route-specific templates that extend a layout.

**pages/dashboard.tmpl**
```html
{{define "title"}}Dashboard{{end}}

{{define "content"}}
    <h1>Test Dashboard</h1>
    {{template "test_list" .}}
{{end}}

{{define "scripts"}}
    <script src="/static/js/dashboard.js"></script>
{{end}}

{{template "layouts/base" .}}
```

### 3. Partials
Reusable components included by pages and layouts.

**partials/_test_card.tmpl**
```html
{{define "test_card"}}
<div class="test-card {{if .Passed}}passed{{else}}failed{{end}}">
    <h3>{{.Name}}</h3>
    <p>Duration: {{.Duration}}</p>
    {{if .Error}}
        <pre class="error">{{.Error}}</pre>
    {{end}}
</div>
{{end}}
```

## 🎭 Template Functions

### Built-in Functions
- `eq`, `ne`, `lt`, `gt`, `le`, `ge` - Comparison functions
- `index` - Access map or array elements
- `len` - Get length of array, slice, or map
- `printf` - Formatted printing
- `html`, `js` - Context-aware escaping

### Custom Functions
- `formatTime` - Format time in a human-readable format
- `truncate` - Truncate text with ellipsis
- `classIf` - Conditionally add CSS classes
- `icon` - Render an icon from the icon system

## 🛠 Development

### Template Inheritance

1. **Base Template**: Define blocks that can be overridden by child templates
2. **Page Template**: Extend base template and implement required blocks
3. **Partials**: Include reusable components using `{{template "_partial_name" .}}`

### Best Practices

1. **Keep Logic Minimal**: Move complex logic to handlers
2. **Use Partials**: Break down large templates into smaller components
3. **Consistent Naming**: Use `snake_case.tmpl` for file names
4. **Document Blocks**: Comment complex template sections
5. **Security**: Always use context-aware escaping (`{{.Var}}` not `{{.}}`)

## 🧪 Testing

### Unit Testing
Test template rendering with different data scenarios:

```go
func TestDashboardTemplate(t *testing.T) {
    tpl := template.Must(template.ParseFiles("layouts/base.tmpl", "pages/dashboard.tmpl"))
    
    data := struct {
        Title string
        Tests []TestResult
    }{
        Title: "Test Dashboard",
        Tests: []TestResult{...},
    }
    
    var buf bytes.Buffer
    if err := tpl.Execute(&buf, data); err != nil {
        t.Fatalf("Template execution failed: %v", err)
    }
    
    // Assertions on buf.String()
}
```

### Linting
Use `templ` to lint and validate templates:

```bash
# Check for common errors
templ check ./...

# Format templates
templ fmt -w .
```

## 🔍 Debugging

1. **Template Errors**: Check for missing blocks or variables
2. **Whitespace Control**: Use `{{-` and `-}}` to trim whitespace
3. **Context**: Ensure proper data is passed to templates
4. **Inheritance**: Verify template lookup paths

## 📚 Resources

- [Go Template Documentation](https://pkg.go.dev/text/template)
- [HTMX Documentation](https://htmx.org/docs/)
- [Go Template Cookbook](https://github.com/benbjohnson/tmpl)

## 🤝 Contributing

1. Follow the template hierarchy and naming conventions
2. Add tests for new templates
3. Document any new template functions or variables
4. Keep templates focused and maintainable

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](../../../LICENSE) file for details.
