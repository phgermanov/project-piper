#!/usr/bin/env node
'use strict';
var express = require('express');
var app = express();

const PORT = 80;
const HOST = '0.0.0.0';

var home = require('./backend/home.js');
var getData = require('./backend/dataMgmt.js');
var getProcessingStatus = require('./backend/status.js');

app.get('/', home);

app.get('/data', getData);

app.get('/status', getProcessingStatus);

app.post('/', home);

app.use(express.static('src/frontend'));


module.exports = app.listen(PORT, HOST, function () {
    console.log('server starting on ' + `http://${HOST}:${PORT}`); 
});;
