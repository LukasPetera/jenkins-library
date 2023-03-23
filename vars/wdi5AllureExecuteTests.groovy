import static com.sap.piper.Prerequisites.checkScript

import com.sap.piper.ConfigurationHelper
import com.sap.piper.GenerateDocumentation
import com.sap.piper.GitUtils
import com.sap.piper.Utils
import com.sap.piper.k8s.ContainerMap
import groovy.transform.Field
import groovy.text.GStringTemplateEngine

@Field String STEP_NAME = getClass().getName()

//TODO: limit parameter visibility
@Field Set GENERAL_CONFIG_KEYS = [
    /**
     * Defines the tool which is used for executing the tests
     * @possibleValues `maven`, `npm`, `bundler`
     */
    'buildTool',
    /** @see dockerExecute */
    'dockerEnvVars',
    /** @see dockerExecute */
    'dockerImage',
    /** @see dockerExecute */
    'dockerName',
    /** @see dockerExecute */
    'dockerOptions',
    /** @see dockerExecute */
    'dockerWorkspace',
    /**
     * With `failOnError` the behavior in case tests fail can be defined.
     * @possibleValues `true`, `false`
     */
    'failOnError',
    /**
     * Only if `testRepository` is provided: Branch of testRepository, defaults to master.
     */
    'gitBranch',
    /**
     * Only if `testRepository` is provided: Credentials for a protected testRepository
     * @possibleValues Jenkins credentials id
     */
    'stashContent',
    /**
     * Define an additional repository where the test implementation is located.
     * For protected repositories the `testRepository` needs to contain the ssh git url.
     */
    'testRepository'
]
@Field Set STEP_CONFIG_KEYS = GENERAL_CONFIG_KEYS
@Field Set PARAMETER_KEYS = STEP_CONFIG_KEYS

/**
 * TODO: Documentdation
 */
@GenerateDocumentation
void call(Map parameters = [:], Closure body) {
    //echo "XXXXXXXXXXXXXX_1"
    //echo 'XXXXXXXXXXXXXX_1'
    sh "echo XXXX1"

    handlePipelineStepErrors(stepName: STEP_NAME, stepParameters: parameters) {
        def script = checkScript(this, parameters) ?: this
        def utils = parameters?.juStabUtils ?: new Utils()
        String stageName = parameters.stageName ?: env.STAGE_NAME

        //echo "XXXXXXXXXXXXXX_2"
        //echo 'XXXXXXXXXXXXXX_2'
        //sh returnStatus: true, script: "echo SHXXXXXXXXXXXXXX_2"
        //sh returnStatus: true, script: 'echo SHXXXXXXXXXXXXXX_2'
        sh "echo XXXX2"

        // load default & individual configuration
        Map config = ConfigurationHelper.newInstance(this)
            .loadStepDefaults([:], stageName)
            .mixinGeneralConfig(script.commonPipelineEnvironment, GENERAL_CONFIG_KEYS)
            .mixinStepConfig(script.commonPipelineEnvironment, STEP_CONFIG_KEYS)
            .mixinStageConfig(script.commonPipelineEnvironment, stageName, STEP_CONFIG_KEYS)
            .mixin(parameters, PARAMETER_KEYS)
            .dependingOn('buildTool').mixin('dockerImage')
            .dependingOn('buildTool').mixin('dockerName')
            .dependingOn('buildTool').mixin('dockerWorkspace')
            .use()

        utils.pushToSWA([
            step: STEP_NAME,
            stepParamKey1: 'scriptMissing',
            stepParam1: parameters?.script == null
        ], config)

        // Inject config via env vars so that scripts running inside selenium can respond to that
        config.dockerEnvVars = config.dockerEnvVars ?: [:]

        dockerExecute(
                script: script,
                dockerEnvVars: config.dockerEnvVars,
                dockerImage: config.dockerImage,
                dockerName: config.dockerName,
                dockerOptions: config.dockerOptions,
                dockerWorkspace: config.dockerWorkspace
        ) {
            try {
                sh returnStatus: true, script: """
                    node --version
                    npm --version
                """
                config.stashContent = config.testRepository
                    ?[GitUtils.handleTestRepository(this, config)]
                    :utils.unstashAll(config.stashContent)
      
                    //TODO: Execute WDI5 Tests & Allure reports
                    //echo "XXXXXXXXXXXXXX_3"
                    //sh returnStatus: true, script: "echo SHXXXXXXXXXXXXXX_3"
                    //echo 'XXXXXXXXXXXXXX_3'
                    //sh returnStatus: true, script: 'echo SHXXXXXXXXXXXXXX_3'
                        sh "echo XXXX3"

//                     sh 'node_modules/.bin/wdio run wdio.conf.js --spec test/dataPreparation.js --tester_user=${username_tester} --tester_pass=${password_tester} --baseUrl $BASEURL'
//                     sh 'node_modules/.bin/wdio run wdio.conf.js --tester_user=${username_tester} --tester_pass=${password_tester} --baseUrl $BASEURL'
                
            } catch (err) {
                if (config.failOnError) {
                    throw err
                }
            }
        }
    }
}
