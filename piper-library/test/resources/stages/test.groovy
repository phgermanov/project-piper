void call(body, stageName, config) {
    echo "Start - Extension for stage: ${stageName}"
    echo "Current stage config: ${config}"
    body()
    echo "End - Extension for stage: ${stageName}"
}
return this
