# How to enable ESLint (S/4 rules, SAP-static-code-checks)

Use the [gruntify-eslint](https://www.npmjs.com/package/gruntify-eslint) module and provide a grunt target to create the appropriate eslint report.

Example:

``` javascript
eslint : {
  options : {
    format : 'jslint-xml',
    outputFile : 'target/eslint.jslint.xml',
    silent : true
  },
  src : [ 'webapp/' ]
}
```

TODO What are the correct eslint settings to specify in .eslintrc.yml
