package com.sap.icd.jenkins

import groovy.text.GStringTemplateEngine
// TODO: maybe later:
//import groovy.text.XmlTemplateEngine

// import com.cloudbees.groovy.cps.NonCPS

/**
 * Renders a string template(using groovy template engine)
 *
 * @param template the template string
 * @param context the binding context map
 * @return String the resulting rendered string
 */
//@NonCPS
private static String renderImpl(String template, Map context) {
    def engine = new GStringTemplateEngine()
    def tpl = engine.createTemplate(template)
    String result = tpl.make(context).toString()
    // in case underlying class has CPS issues:
    template = null
    engine = null
    return result
}

/**
 * Renders (lazily) template variable using context
 *
 * @param template
 * @param context
 * @return
 */
public static String render(String template, Map context=[:]) {
    // empty template -> avoid calling renderImpl
    if (!template) { return template }
    // empty context -> avoid calling render
    if (!(context)) { return template }
    String result = renderImpl(template, context)
    return result
}
