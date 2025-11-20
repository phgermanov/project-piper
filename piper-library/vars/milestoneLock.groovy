def call(Map parameters = [:], Closure body) {    
  def ordinal = parameters.get('ordinal', '')    
  def inversePrecedence = parameters.get('inversePrecedence', true)    
  lock(resource: "${env.JOB_NAME}/${ordinal}", inversePrecedence: inversePrecedence) {        
    milestone ordinal        
    body()    
  }
}
