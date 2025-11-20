/**
 * conditionalSteps
 *
 * @param condition
 * @param body
 */
def call(Map parameters = [:], body) {
    if (parameters.condition) {
        body()
    }else{
        echo 'Skipping steps due to not fulfilled condition.'
    }
}
