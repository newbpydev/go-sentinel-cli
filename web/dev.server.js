// Express dev server with Handlebars templates, layouts, partials and WebSockets
const express = require('express');
const exphbs = require('express-handlebars');
const path = require('path');
const http = require('http');
const { setupWebSocketServer } = require('./websocket-server');
const { GoWebSocketAdapter } = require('./server/ws-adapter');

// Configuration for the WebSocket adapter
const WS_BACKEND_PORT = process.env.WS_BACKEND_PORT || 8080; // Go backend WebSocket port

const app = express();
const PORT = process.env.PORT || 5174;

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

// Create HTTP server from Express app
const server = http.createServer(app);

// Setup WebSocket server
const wss = setupWebSocketServer(server);

// Setup WebSocket adapter to connect to Go backend
const wsAdapter = new GoWebSocketAdapter(WS_BACKEND_PORT);
wsAdapter.start(wss);

// Add environment variables to the template context
app.locals.wsBackendPort = WS_BACKEND_PORT;

// Start the server
server.listen(PORT, () => {
  console.log(`Dev server running: http://localhost:${PORT}`);
  console.log(`WebSocket server running at ws://localhost:${PORT}/ws`);
  console.log(`WebSocket adapter connecting to Go backend at ws://localhost:${WS_BACKEND_PORT}/ws`);
});
