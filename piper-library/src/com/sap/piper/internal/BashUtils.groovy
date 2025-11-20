package com.sap.piper.internal

class BashUtils implements Serializable {
    static final long serialVersionUID = 1L

    static String escape(String str, wrapInQuotes = true) {
        // put string in single quotes and escape contained single quotes by putting them into a double quoted string

        def escapedString = str.replace("'", "'\"'\"'")
        if (wrapInQuotes) {
            escapedString = "'${escapedString}'"
        }
        return escapedString
    }
}
