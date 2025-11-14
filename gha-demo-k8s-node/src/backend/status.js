function getProcessingStatus (req, res) { 
    res.status(200);
    res.setHeader('Content-Type', 'application/json');
    res.send(JSON.stringify({
        OrderState: [
                    {
                        product: 'Notebook Basic 15',
                        state: 'received'
                    },
                    {
                        product: 'Notebook Basic 20',
                        price: 'in delivery',
                        trackingCode: 'HGJSG77654645'
                    }
                ]
        }));
}

module.exports = getProcessingStatus;
