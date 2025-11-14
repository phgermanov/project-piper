var gulp = require('gulp');
var mocha = require('gulp-mocha');
var istanbul = require('gulp-istanbul');
var plumber = require('gulp-plumber');
var Server = require('karma').Server;


gulp.task('pre-test', function () {
  return gulp.src(['src/*.js', 'src/backend/**/*.js', '!src/backend/test/*-spec.js'])
    // Covering files
    .pipe(istanbul({
        includeUntested: true
    }))
    .pipe(istanbul.hookRequire());
});


gulp.task('test:unit', gulp.series('pre-test', function(cb) {
    var mochaErr;
    gulp.src(__dirname + '/src/backend/test/*-spec.js', {read: false})
        .pipe(plumber())
        .pipe(mocha({
            reporter: 'mocha-junit-reporter',
            reporterOptions:{
                mochaFile: __dirname + '/target/coverage/TEST-unit.xml'
            }
        }))
        .on('error', (err) => {
            mochaErr = err;
        })
        .pipe(istanbul.writeReports({
            dir: __dirname + '/target/coverage',
            reporters: [ 'text', 'cobertura', 'html', 'lcov' ]
        }))
        .on('end', () => {
            cb(mochaErr);
        });
}));

// opa tests
gulp.task('test:opa', function (done) {
  new Server({
    configFile: __dirname + '/karma.conf.js',
    browsers: ['chrome_selenium']
  }, done).start();
});

gulp.task('test', gulp.series('test:unit', 'test:opa', function() {
}));

gulp.task('default', gulp.series('test:unit', function() {
}));