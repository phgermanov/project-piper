package com.sap.piper.internal

import static java.lang.Boolean.getBoolean

static checkScript(def step, Map params) {
    def msg = "No reference to surrounding script provided with key 'script', e.g. 'script: this'."
    def script = params?.script

    if(script == null) {
        if(getBoolean('com.sap.piper.featureFlag.failOnMissingScript')) {
            Notify.error(step, msg)
        }else{
            Notify.warning(step, "${msg} In future versions of piper-lib the build will fail.")
        }
    }

    return script
}
