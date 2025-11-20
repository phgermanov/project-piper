# Hints for setting up local testing of pipeline library using maven

## SAP Certificates

You can download the certificates here:

* `http://aia.pki.co.sap.com/aia/SAPNetCA_G2.crt`
* `http://aia.pki.co.sap.com/aia/SAP%20Global%20Root%20CA.crt`

You need to store these certificates in your local Java keystore

```sh
keytool -noprompt -import -alias sapglobalrootca -file "SAP Global Root CA.crt" -keystore "C:\SAP\sapjvm_7\jre\lib\security\cacerts" -storepass changeit
keytool -noprompt -import -alias sapnetcag2 -file SAPNetCA_G2.crt  -keystore "C:\SAP\sapjvm_7\jre\lib\security\cacerts" -storepass changeit
```

**Important: Make sure to use the certificate store used by your default JVM in case you have multiple installed**

## Testing inside a Docker container

It is highly recommended to use a Docker container for testing, using e.g. the [jenkins-master](https://github.wdf.sap.corp/IndustryCloudFoundation/jenkins-master) image.

In order to run a test you need to:

* map your directory containing the tests as a volume of the container
* map the docker socket in order to be able to start docker

Example using jenkins-master on a windows machine using *docker-machine*

```sh
docker run --rm -it -v /c/Users/d032835/IdeaProjects/test-jenkins-lib:/var/jenkinslib --entrypoint /bin/bash icf.int.repositories.cloud.sap/icf/jenkins
```

Inside this container you can run the tests using

```sh
mvn clean test
```

In case you want to run a specific test you can run this test using:

```sh
mvn surefire:test -Dtest=<TestClass>
```
