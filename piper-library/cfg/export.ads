artifacts builderVersion:"1.1", {
  version "${buildBaseVersion}", {
    group "com.sap.prd.cicd", {
      artifact "sap-piper", {
        file "${genroot}/tmp/src/DockerLayersExtract/sap-piper", classifier: "${buildRuntime}", extension: "out"
      }
    }
  }
}