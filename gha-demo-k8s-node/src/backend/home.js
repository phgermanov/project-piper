var path = require('path');

function init(req, res) {
    res.status(200);
    res.sendFile(path.join(__dirname + '/../frontend/index.html'));
}

module.exports = init;