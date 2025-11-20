#!groovy

package util

import com.sap.piper.internal.integration.Fortify
import com.sap.piper.internal.JenkinsUtils
import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.integration.ZedAttackProxy
import org.codehaus.groovy.runtime.InvokerHelper
import org.springframework.context.annotation.Bean
import org.springframework.context.annotation.Configuration

@Configuration
class PiperTestConfiguration {

    @Bean
    Script nullScript() {
        def nullScript = InvokerHelper.createScript(null, new Binding())
        nullScript.currentBuild = [:]
        nullScript.env = [:]
        nullScript.STEP_NAME = 'NullScript'
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(nullScript)
        return nullScript
    }

    @Bean
    Utils mockUtils() {
        def mockUtils = new Utils()
        mockUtils.env = [:]
        mockUtils.steps = [
            stash  : { m -> println "stash name = ${m.name}" },
            unstash: { println "unstash called ..." }
        ]
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(mockUtils)
        return mockUtils
    }

    @Bean
    JenkinsUtils mockJenkinsUtils() {
        def mockJenkinsUtils = new JenkinsUtils()
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(mockJenkinsUtils)
        return mockJenkinsUtils
    }

    @Bean
    MockHelper mockHelper() {
        return new MockHelper()
    }

    @Bean
    Fortify mockFortify() {
        def mockFortify = new Fortify(
            [
                fortifyToken: 'ddiuzasiudzasd687686773e',
                fortifyServerUrl: 'https://fortify.mo.sap.corp/ssc',
                fortifyApiEndpoint: '/api/v1',
                verbose: true,
                buildDescriptorFile: './pom.xml',
                fortifyFprUploadEndpoint: '/upload/resultFileUpload.html',
                fortifyFprDownloadEndpoint: '/download/currentStateFprDownload.html',
                fortifyReportDownloadEndpoint: '/transfer/reportDownload.html'
            ]
            , mockUtils()
        )
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(mockFortify)
        return mockFortify
    }

    @Bean
    ZedAttackProxy mockZedAttackProxy() {
        def mockZedAttackProxy = new ZedAttackProxy(nullScript(), true)
        LibraryLoadingTestExecutionListener.prepareObjectInterceptors(mockZedAttackProxy)
        return mockZedAttackProxy
    }
}
