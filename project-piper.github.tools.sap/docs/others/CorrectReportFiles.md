# Correct Files Paths in Check/Test Result Files

Some tools like *CPD (code duplication)*, *PMD (static checks)* or *LCOV (code coverage)* that create report files use absolute paths. While this is ok for local execution, it results into issues if the build is done on xMake.

So for example an LCOV report generated on xMake specifies a covered file like this:

```properties
...
SF:/data/xmake/prod-build10010/w/myOrg/myOrg-myRepo-SP-MS-linuxx86_64/gen/out/module/lib/deploy.js
FN:6,(anonymous_0)
FN:16,(anonymous_1)
FN:28,(anonymous_2)
...
```

If you now publish this LCOV report to Sonar, you will get the following warning:

```properties
INFO: Sensor SonarJS [javascript] (done) | time=709ms
INFO: Sensor SonarJS Coverage [javascript]
INFO: Analysing [/home/jenkins/workspace/myRepo-IPS3UUSRUSOMYG3ZJW33BFH7UUHBRP4MV7INRYCMAN7QJ2IBKLGQ/target/coverage/lcov.info]
INFO: 6/6 source files have been analyzed
WARN: Could not resolve 7 file paths in [/home/jenkins/workspace/myRepo-IPS3UUSRUSOMYG3ZJW33BFH7UUHBRP4MV7INRYCMAN7QJ2IBKLGQ/target/coverage/lcov.info], first unresolved path: /data/xmake/prod-build10010/w/myOrg/myOrg-myRepo-SP-MS-linuxx86_64/gen/out/module/lib/deploy.js
INFO: Sensor SonarJS Coverage [javascript] (done) | time=5ms
```

A way to overcome this, is to rewrite the report files and remove the absolute paths and replace them  with a relative one. This can be done by using the shell commands `find` and `sed` to alter all report files before publishing them to Sonar.

```groovy
// removing the xmake file prefix
def reportFile = 'lcov.info'
def search = '\\/data\\/xmake\\/.*\\/gen\\/out\\/module\\/'

try {
    sh "find . -type f -name '${reportFile}' -exec sed -i 's/${search}//g' {} \\;"
} catch (ignore) {}
```

Make sure you have the correct search pattern defined for your use case.
