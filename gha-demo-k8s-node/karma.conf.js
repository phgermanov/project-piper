// karma.conf.js
module.exports = function(config) {
  config.set({
    basePath : "./",
        //proxies: {
        //},
        frameworks : [ "qunit", "openui5" ],
        files : [       
            {
                pattern : "src/frontend/test/karma-qunit.js", included: true
            },
            {
                pattern : "src/frontend/test/integration/AllJourneys.js", included: true
            },
            {
                pattern : "src/frontend/**/*", included : false
            }
        ],
        // test results reporter to use
        // possible values: 'dots', 'progress'
        // available reporters: https://npmjs.org/browse/keyword/karma-reporter
        reporters: ['progress', 'coverage', 'junit'],
        // preprocess matching files before serving them to the browser
        // available preprocessors: https://npmjs.org/browse/keyword/karma-preprocessor
        preprocessors: {
            'src/frontend/*.js': ['coverage'],
            'src/frontend/!(test|localService)/**/*.js': ['coverage']
        },
        hostname: 'karma',
        customLaunchers: {
            chrome_selenium: {
                base: 'WebDriver',
                config: {
                     hostname: 'selenium',
                     port: 4444
                },
                browserName: 'chrome',
                name: 'Chrome'
            },
        },
        port : 9876,
        logLevel : "DEBUG",
        autoWatch : false,
        singleRun : true,
        browserNoActivityTimeout : 40000,
        browserDisconnectTolerance: 2,

        openui5 : {
            path : "https://sapui5.hana.ondemand.com/resources/sap-ui-core.js"
        },

        junitReporter: {
            outputDir: 'target/karma', 
            outputFile: 'TEST-qunit.junit.xml', 
            suite: '', // suite will become the package name attribute in xml testsuite element
            useBrowserName: true, // add browser name to report and classes names
            nameFormatter: undefined, // function (browser, result) to customize the name attribute in xml testcase element
            classNameFormatter: undefined, // function (browser, result) to customize the classname attribute in xml testcase element
            properties: {} // key value pair of properties to add to the <properties> section of the report
        },
        coverageReporter: {
            type: "cobertura",
            dir: "target/karma",
            includeAllSources: true,
            check: {
                global: {
                    lines: 60
                }
            }
        },

        client: {
            openui5 : {
                config : {
                    theme : "sap_bluecrystal",
                    resourceroots : {
                    'sap.ui.piper.test': '/base/src/frontend/test',
                    'sap.ui.piper.controller': '/base/src/frontend/controller',
                    'sap.ui.piper': '/base/src/frontend'
                    }
                }
            },
        }
    }   
  );

};