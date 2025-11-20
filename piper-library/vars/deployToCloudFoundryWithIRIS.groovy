import com.sap.icd.jenkins.Utils
import com.sap.piper.internal.ConfigurationHelper
import com.sap.piper.internal.Deprecate
import groovy.transform.Field

import static com.sap.piper.internal.Prerequisites.checkScript

@Field String STEP_NAME = 'deployToCloudFoundryWithIRIS'
@Field Set GENERAL_CONFIG_KEYS = [
    'run',
    'cfAppName',
    'cfManifest',
    'verbose',
    'transferFile'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

//////////////////////////////////////////////////////////////////////
//                                                                  //
//   IRIS - Immortal Repository for Integrated operations Scripts   //
//                                                                  //
//////////////////////////////////////////////////////////////////////

def call(Map parameters = [:]) {
    handlePipelineStepErrors (stepName: STEP_NAME, stepParameters: parameters,
        libraryDocumentationUrl: 'https://go.sap.corp/piper/',
        libraryRepositoryUrl: 'https://github.wdf.sap.corp/ContinuousDelivery/piper-library/'
    ) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters.juStabUtils ?: new Utils()
        // handle deprecated parameters
        Deprecate.parameter(this, parameters, 'transferfile', 'transferFile')
        // load default & individual configuration
        Map config = ConfigurationHelper
            .loadStepDefaults(this)
            .mixinGeneralConfig(script.globalPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.globalPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.globalPipelineEnvironment, parameters.stageName?:env.STAGE_NAME, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .use()

        config = new ConfigurationHelper(config)
            .mixin([
                run: config.run instanceof Boolean ? config.run : config.run.toBoolean()
            ])
            .use()

        if (config.run) {
            if (fileExists(config.transferFile)) {
                def manifestRepack = repackManifest(config.cfManifest, config.verbose)

                def systems = readJSON file: config.transferFile
                systems.each{
                    def cfAppName = config.cfAppName
                    def cfApiEndpoint = it.value.cfApiEndpoint.toString()
                    def cfOrg = it.value.cfOrg.toString()
                    def cfSpace = it.value.cfSpace.toString()
                    def cfManifest = ''
                    def deployMessage = ''

                    // Identify matching manifest
                    if ( manifestRepack[cfApiEndpoint + "|" + cfOrg + "|" + cfSpace] ) {
                        cfManifest = manifestRepack[cfApiEndpoint + "|" + cfOrg + "|" + cfSpace]
                    } else {
                        if ( manifestRepack[cfApiEndpoint + "|" + cfOrg] ) {
                            cfManifest = manifestRepack[cfApiEndpoint + "|" + cfOrg]
                        } else {
                            if ( manifestRepack[cfApiEndpoint] ) {
                                cfManifest = manifestRepack[cfApiEndpoint]
                            } else {
                                cfManifest = manifestRepack['else']
                            }
                        }
                    }

                    // Read cfAppName from manifest
                    if (cfAppName == null || cfAppName == '') {
                        if (fileExists(cfManifest)) {
                            def manifest = readYaml file: cfManifest

                            if (manifest && manifest.applications && manifest.applications[0] && manifest.applications[0].name) {
                                cfAppName = manifest.applications[0].name
                                deployMessage = " (AppName taken from Manifest)"
                            }
                        }
                    }

                    // Deploy
                    echo "[${STEP_NAME}] Deploy ${it.value.SystemRoleCode} with AppName \"${cfAppName}\" to Org \"${cfOrg}\" / Space \"${cfSpace}\" in DataCenter ${it.value.DataCenter} (${cfApiEndpoint}) with Manifest '${cfManifest}'${deployMessage}"

                    deployToCloudFoundry parameters + [script: script, cfAppName: cfAppName, cfApiEndpoint: cfApiEndpoint, cfOrg: cfOrg, cfSpace: cfSpace, cfManifest: cfManifest]
                }
            } else {
                echo "[${STEP_NAME}] Transferfile ${config.transferFile} does not exist. Nothing to deploy!"
            }
        } else {
            echo 'Don\'t execute, based on configuration setting.'
        }
    }
}

// Repack for easier parsing
private Map repackManifest(String cfManifest, verbose) {
  return repackManifest([1: [cfManifest: cfManifest]], verbose)
}

// Repack for easier parsing
private Map repackManifest(Map cfManifest = [:], verbose) {
  def manifestRepack = [:]
  manifestRepack['else'] = 'manifest.yml'

  cfManifest.each{
      if ( it.value.cfApiEndpoint && it.value.cfOrg && it.value.cfSpace) {
          manifestRepack[it.value.cfApiEndpoint.toString() + "|" + it.value.cfOrg.toString() + "|" + it.value.cfSpace.toString()] = it.value.cfManifest.toString()
      } else  {
          if ( it.value.cfApiEndpoint && it.value.cfOrg && !it.value.cfSpace) {
              manifestRepack[it.value.cfApiEndpoint.toString() + "|" + it.value.cfOrg.toString()] = it.value.cfManifest.toString()
          } else  {
              if ( it.value.cfApiEndpoint && !it.value.cfOrg && !it.value.cfSpace) {
                  manifestRepack[it.value.cfApiEndpoint.toString()] = it.value.cfManifest.toString()
              } else  {
                  manifestRepack['else'] = it.value.cfManifest.toString()
              }
          }
      }
  }
  if ( verbose ) {
      echo "cfManifest before repack = " + cfManifest
      echo "cfManifest after repack = " + manifestRepack
  }

  return manifestRepack
}
