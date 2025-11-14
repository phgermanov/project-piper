# How to enable UI Test Code Coverage

When using Karma for executing unit tests, the code coverage can be collected using a preprocessor.

``` javascript
preprocessors: {
  'webapp/*.js': ['coverage'],
  'webapp/!(test|localService)/**/*.js': ['coverage']
}

...

reporters: ['progress', 'coverage', 'junit'],

...

coverageReporter: {
  // specify a common output directory
  dir: 'target/coverage',
  includeAllSources: true,
  reporters: [{
    type: 'cobertura',
    file: 'cobertura-coverage.xml'
  }, {
    type: 'lcovonly',
    file: 'lcov-coverage.txt'
  }]
}
```

Note: Code coverage for OPA5 tests is not measured when using IFrames.
