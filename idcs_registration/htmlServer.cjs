var express = require('express'),
  path = require('path'),
  cors = require('cors'),
  app = express();

// set the port
app.set('port', 8000);

// enable cors for all origins
app.use(cors())

// tell express that we want to use the dist folder
// for our static assets
app.use(express.static(path.join(__dirname, '/dist/Html')));

// Listen for requests
var server = app.listen(app.get('port'), function () {
  console.log('The server is running on http://localhost:' + app.get('port'));
});