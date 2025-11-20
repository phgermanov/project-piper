package com.sap.piper.internal

class MapUtils implements Serializable {
    static isMap(object){
        return object in Map
    }

    static Map pruneNulls(Map m) {
        Map result = [:]
        m = m ?: [:]
        m.each { key, value ->
            if(isMap(value))
                result[key] = pruneNulls(value)
            else if(value != null)
                result[key] = value
        }
        return result
    }

    static Map merge(Map base, Map overlay) {
        Map result = [:]
        base = base ?: [:]
        result.putAll(base)
        overlay.each { key, value ->
            result[key] = isMap(value) ? merge(base[key], value) : value
        }
        return result
    }

    static Map deepcopy(Map orig) {
        ByteArrayOutputStream bos = new ByteArrayOutputStream()
        ObjectOutputStream oos = new ObjectOutputStream(bos)
        oos.writeObject(orig); oos.flush()
        ByteArrayInputStream bin = new ByteArrayInputStream(bos.toByteArray())
        ObjectInputStream ois = new ObjectInputStream(bin)
        return ois.readObject()
    }
}
