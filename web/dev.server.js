// Express dev server with Handlebars templates, layouts, and partials
const express = require('express');
const exphbs = require('express-handlebars');
const path = require('path');

const app = express();
const PORT = process.env.PORT || 5173;

// Handlebars engine setup
app.engine('hbs', exphbs.engine({
  extname: '.hbs',
  defaultLayout: 'base',
  layoutsDir: path.join(__dirname, 'templates', 'layouts'),
  partialsDir: path.join(__dirname, 'templates', 'partials'),
}));
app.set('view engine', 'hbs');
app.set('views', path.join(__dirname, 'templates', 'pages'));

// Serve static files
app.use('/static', express.static(path.join(__dirname, 'static')));

// Add middleware to log template rendering for debugging
app.use((req, res, next) => {
  const originalRender = res.render;
  res.render = function(view, options, callback) {
    console.log(`Rendering template: ${view} with Handlebars`);
    originalRender.call(this, view, options, callback);
  };
  next();
});

// Static test page for direct Tailwind testing
app.get('/test', (req, res) => {
  res.sendFile(path.join(__dirname, 'templates', 'test.html'));
});

// Index route (dashboard)
app.get(['/', '/index.html'], (req, res) => {
  res.render('index', {});
});

// Fallback route
app.use((req, res) => {
  res.status(404).send('Not Found');
});

app.listen(PORT, () => {
  console.log(`Dev server running: http://localhost:${PORT}`);
});
