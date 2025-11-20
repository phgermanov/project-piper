package com.sap.icd.jenkins

class CredentialsWrapper {

    private String credentialsId
    private Script script

    public CredentialsWrapper(Script script,String credentialsId) {
        this.script = script
        this.credentialsId=credentialsId
    }

    public int asCFCredentialsEncoded(Closure body) {
        script.withCredentials([
            script.usernamePassword(credentialsId: credentialsId, passwordVariable: 'CF_PASSWORD_STORE', usernameVariable: 'CF_USERNAME')
        ]) {
            script.withEnv(['CF_PASSWORD_ENCODED=true',"CF_PASSWORD=${encodePassword(script.CF_PASSWORD_STORE)}"]) {
                return body()
            }
        }
    }

    public int asMuenchhausenCredentialsEncoded(Closure body) {
        if (credentialsId) {
            script.withCredentials([
                script.usernamePassword(credentialsId: credentialsId, passwordVariable: 'MH_PASSWORD_STORE', usernameVariable: 'MH_USERNAME')
            ]) {
                script.withEnv(['MH_PASSWORD_ENCODED=true',"MH_PASSWORD=${encodePassword(script.MH_PASSWORD_STORE)}"]) {
                   return body()
                }
            }
        }
        else {
            return body()
        }
    }

    private String encodePassword(String password) {
        return (password.trim() + '\n').bytes.encodeBase64().toString()
    }
}
