function authenticate(helper, paramsValues, credentials) {
    print('Authenticating via oAuth client credentials flow');

    var authHelper = new OAuthAuthenticator(helper, paramsValues, credentials);

    return authHelper.login();
}

function getRequiredParamsNames() {
    return ["API URL"];
}

function getOptionalParamsNames() {
    return ["grant_type", "response_type"];
}

function getCredentialsParamsNames() {
    return ["client_id", "client_secret"];
}

function OAuthAuthenticator(helper, paramsValues, credentials) {

    this.helper = helper;
    this.loginApiUrl = paramsValues.get('API URL');
    this.clientId = credentials.getParam('client_id');
    this.clientSecret = credentials.getParam('client_secret');
    this.grant_type = credentials.getParam('grant_type');
    if(this.grant_type == null)
	  this.grant_type = 'client_credentials';
    if(this.response_type == null)
	  this.response_type = 'token';
     
    return this;
}

OAuthAuthenticator.prototype = {
    login: function () {
        var HttpRequestHeader = Java.type('org.parosproxy.paros.network.HttpRequestHeader'),
            requestBody = 'client_id=' + encodeURIComponent(this.clientId) + '&client_secret=' + encodeURIComponent(this.clientSecret) + '&grant_type=' + encodeURIComponent(this.grant_type) + '&response_type=' + encodeURIComponent(this.response_type);
            
        var response = this.doRequest(
                this.loginApiUrl,
                HttpRequestHeader.POST,
                requestBody
            );
	   var rawResponse = response.getResponseBody().toString();
        var parsedResponse = JSON.parse(rawResponse);
        
        if (parsedResponse.error == 'unauthorized' || parsedResponse.access_token == undefined) {
            print('Authentication failure to ' + this.loginApiUrl + ' with : ' + requestBody + " : " + parsedResponse.error_description);
        }
        else {
	       var token = parsedResponse.access_token
            print('Authentication success. Token = ' + token); 
            org.zaproxy.zap.extension.script.ScriptVars.setGlobalVar('logintoken', token);
        }
        return response;
    },

    doRequest: function (url, requestMethod, requestBody) {
        var HttpRequestHeader = Java.type('org.parosproxy.paros.network.HttpRequestHeader'),
            HttpHeader = Java.type('org.parosproxy.paros.network.HttpHeader'),
            URI = Java.type('org.apache.commons.httpclient.URI'),
            msg,
            requestUri = new URI(url, false),
            requestHeader = new HttpRequestHeader(requestMethod, requestUri, HttpHeader.HTTP10);

        msg = this.helper.prepareMessage();
        requestHeader.setContentLength(requestBody.length);
        msg.setRequestHeader(requestHeader);
        msg.setRequestBody(requestBody);

        this.helper.sendAndReceive(msg);

        return msg;
    }
};
