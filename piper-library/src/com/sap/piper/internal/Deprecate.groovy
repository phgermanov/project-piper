package com.sap.piper.internal

class Deprecate implements Serializable {
    static def parameter(Script step, Map map, String name, String replacement, String type = null) {
        if (map?.get(name) != null) {
            def msg = "The ${type?"${type} ":''}parameter ${name} is DEPRECATED"
            if(replacement != null) {
                msg += ", use ${replacement} instead"
                if(map?.get(replacement) == null)
                    map.put(replacement, map?.get(name))
            }
            msg += '!'
            Notify.warning(step, msg)
        }
    }

    static def value(Script step, Map map, String name, deprecatedValue, replacementValue = null) {
        if (map?.get(name) == deprecatedValue) {
            def msg = "The value '${deprecatedValue}' for the parameter ${name} is DEPRECATED"
            if(replacementValue) {
                msg += ", use '${replacementValue}' instead"
                map.put(name, replacementValue)
            }
            msg += '!'
            Notify.warning(step, msg)
        }
    }
}
