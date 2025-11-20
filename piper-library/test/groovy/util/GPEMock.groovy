#!groovy
package util

/**
 * This is a Helper class to load test data for mocks.
 *
 **/

class GPEMock{
    GPEMock() { reset() }
    def artifactVersion
    def influxStepData
    def configuration
    def configProps
    def xMakeProps
    def appContainerProperties
    def influxCustomData
    def githubOrg
    def githubRepo
    def gitBranch
    def gitSshUrl
    def gitHttpsUrl

    def setInfluxStepData(key, value) {
        this.influxStepData[key] = value
    }
    def getConfigProperty(property) {
        return configProps[property]
    }
    def setConfigProperty(property, value) {
        this.configProps[property] = value
    }
    def setXMakeProperties(Map properties) {
        xMakeProps = properties
    }
    def getXMakeProperty(property) {
        return xMakeProps[property]
    }
    def setAppContainerProperty(property, value) {
        appContainerProperties[property] = value
    }
    def setInfluxCustomDataProperty(property, value) {
        influxCustomData[property] = value
    }
    def getConfigPropertyAsBoolean(property) {
        def value = getConfigProperty(property)
        return value == null?false:value.toBoolean()
    }

    def reset() {
        artifactVersion = '1.2.3-20180101-010203_0f54a5d53bcd29b4d747d8d168f52f2ceddf7198'
        influxStepData = [:]
        configuration = [:]
        configProps = [:]
        xMakeProps = [:]
        appContainerProperties = [:]
        influxCustomData = [:]
        githubOrg = 'testOrg'
        githubRepo = 'testRepo'
        gitBranch = null
        gitSshUrl = null
        gitHttpsUrl = null
    }
}
