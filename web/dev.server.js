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
