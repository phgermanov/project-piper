function sendingRequest(msg, initiator, helper) {
    var loginToken = org.zaproxy.zap.extension.script.ScriptVars.getGlobalVar("logintoken");
    var url = msg.getRequestHeader().getURI().toString();
	if(url.indexOf('oauth') < 0) {
		print('Login token is ' + loginToken)
		
	     // set a Http Header
	      var httpRequestHeader = msg.getRequestHeader();
	      httpRequestHeader.setHeader("Authorization", "Bearer " + loginToken);
	      msg.setRequestHeader(httpRequestHeader);
	
		print('sendingRequest called for url=' + url);
	}
}

function responseReceived(msg, initiator, helper) {
	print('responseReceived called for url=' + msg.getRequestHeader().getURI().toString());
}