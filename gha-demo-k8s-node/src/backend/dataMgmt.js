function getData (req, res) { 
    res.status(200);
    res.setHeader('Content-Type', 'application/json');
    res.send(JSON.stringify({
        Products: [
                    {
                        name: 'Notebook Basic 15',
                        price: 520
                    },
                    {
                        name: 'Notebook Basic 20',
                        price: 750
                    }
                ]
        }));
}

module.exports = getData;