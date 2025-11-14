var mocha = require('mocha');
var chai = require('chai');
var sinon = require('sinon');
var expect = chai.expect;
var assert = sinon.assert;

describe('home', function () {

    var getProcessingStatus;

    beforeEach(function() {
        getProcessingStatus = require(__dirname + '/../status');
    });

    afterEach(function() {
        getProcessingStatus = null;
    });

    it('should return ok code and file path', function() {
        let res = {
            status: sinon.spy(),
            send: sinon.spy(),
            setHeader: sinon.spy()
        };

        getProcessingStatus(null, res);
        
        assert.calledOnce(res.status);
        assert.calledWith(res.status, 200);
        assert.calledOnce(res.setHeader);
        assert.calledWith(res.setHeader, 'Content-Type', 'application/json');
        assert.calledOnce(res.send);

    });

});
