def call(Map parameters = [:], Closure body) {    
    def name = parameters.get('name', '')    
    if (''.equals(name)) {        
        node {            
            executeWork(body)        
        }    
    } else {        
        node (name) {            
            executeWork(body)        
        }    
    }
}
def executeWork(Closure body) {    
    deleteDir()    
    body()
}
