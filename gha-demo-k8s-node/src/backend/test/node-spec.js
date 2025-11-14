var request = require('supertest');
var mocha = require('mocha');
var chai = require('chai');
var sinon = require('sinon');
var fs = require('fs');
var expect = chai.expect;
var assert = sinon.assert;

describe('express', function () {

    var server;

    beforeEach(function() {
        server = require(__dirname + '/../../app');
    });

    afterEach(function() {
        server.close();
    });

    it('responds to get /', function testSlash(done) {
        request(server)
            .get('/')
            .expect(200, fs.readFileSync('src/frontend/index.html', 'utf8') , done);
    });
    it('responds to post /', function testSlash(done) {
        request(server)
            .post('/')
            .expect(200, fs.readFileSync('src/frontend/index.html', 'utf8') , done);
    });

    it('404 everything else', function testPath(done) {
        request(server)
        .get('/non/exiting/path')
        .expect(404, done);
    });

    it('should return json data', function(done) {
        request(server)
            .get('/data')
            .set('Accept', 'application/json')
            .expect('Content-Type', /json/)
            .expect(200, {
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
            }, done);
    });

});