package cmd

import (
	"archive/zip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

const TYPE_DEF int = 0
const TYPE_ALL int = 1
const ALL string = "average"
const ENV_STR string = "env set: "
const RUN_CMD_STR string = "executed runCommand: "
const RUN_START string = "bash -c '/opt/testrunner/start'"
const RUN_START_ALL string = "bash -c '/opt/testrunner/startrunner'"
const RUN_START_JS string = "bash -c '/opt/testrunner/startjs'"
const PARAM_SUPA_CFG string = "ENV_SUPA_CFG"
const CUMULUS_DATA_FILE string = "data.json"
const CUMULUS_SOURCE_FILE string = "supa_upload_cumulus"
const REPO_NOT_FOUND_MESSAGE string = "Valid Github Repo not defined, use local mount"
const SUPA_RESULT_PACKAGE string = "supa_result_data.zip"

type supaSettings struct {
	workingDir string
	resultDir  string
	logDir     string
}

var (
	supa supaSettings
)

// TaskReportData encapsulates information about an executed SUPA task.
type TaskReportData struct {
	ipaProject   string
	ipaScenario  string
	ipaVariant   string
	ipaID        string
	ipaURL       string
	state        string
	cpuTime      string
	respTime     string
	nrStatement  string
	nrRoundTrips string
}

type sapSUPAExecuteTestsUtils interface {
	command.ExecRunner
	//	command.ShellRunner

	piperutils.FileUtils
}

type sapSUPAExecuteTestsUtilsBundle struct {
	*command.Command
	*piperutils.Files
}

func newSapSUPAExecuteTestsUtils() *sapSUPAExecuteTestsUtilsBundle {
	utils := sapSUPAExecuteTestsUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func sapSUPAExecuteTests(config sapSUPAExecuteTestsOptions, telemetryData *telemetry.CustomData) {

	utils := newSapSUPAExecuteTestsUtils()

	supa = supaSettings{
		workingDir: "/tools/SUPA",
		resultDir:  "/home/ubuntu/supaData/results",
		logDir:     "/home/ubuntu/supaData/logs",
	}

	err := runSapSUPAExecuteTests(config, telemetryData, utils, "")
	if err != nil {
		log.Entry().WithError(err).Fatal("sapSUPAExecuteTests execution failed")
	}
	log.Entry().Infoln("sapSUPAExecuteTests execution was successful")
}

func runSapSUPAExecuteTests(config sapSUPAExecuteTestsOptions, telemetryData *telemetry.CustomData, utils sapSUPAExecuteTestsUtils, workdir string) error {

	type_msg := "start test type: "
	var errNoTestType error
	var errTestFailed error
	stype := config.ScriptType
	log.Entry().Info("SUPA Piper step sapSUPAExecuteTests version 1.0 : " + stype)
	switch stype {
	case "all":
		log.Entry().Info(type_msg + stype)
		errTestFailed = runAllPerformanceSingleUserTest(&config, telemetryData, utils)
	case "selenium":
		log.Entry().Info(type_msg + stype)
		errTestFailed = runSeleniumPerformanceSingleUserTest(&config, telemetryData, utils)
	case "wdio":
		log.Entry().Info(type_msg + stype)
		errTestFailed = runWdioPerformanceSingleUserTest(&config, telemetryData, utils)
	case "qmate":
		log.Entry().Info(type_msg + stype)
		errTestFailed = runQmatePerformanceSingleUserTest(&config, telemetryData, utils)
	case "npm":
		log.Entry().Info(type_msg + stype)
		errTestFailed = runNpmPerformanceSingleUserTest(&config, telemetryData, utils)
	case "uiveri5":
		log.Entry().Info(type_msg + stype)
		errTestFailed = runUiveri5PerformanceSingleUserTest(&config, telemetryData, utils)
	case "jmeter":
		log.Entry().Info(type_msg + stype)
		errTestFailed = runJMeterPerformanceSingleUserTest(&config, telemetryData, utils)
	case "krypton":
		log.Entry().Info(type_msg + stype)
		errTestFailed = runKryptonPerformanceSingleUserTest(&config, telemetryData, utils)
	case "any":
		log.Entry().Info(type_msg + stype)
		errTestFailed = runAnyTestInContainer(&config, telemetryData, utils)
	default:
		errNoTestType = errors.New("No valid test type configured: " + stype)
	}
	if errNoTestType != nil {
		log.Entry().WithError(errNoTestType).Fatal("sapSUPAExecuteTests scriptType error")
		return errNoTestType
	}

	//archive report even if tests failed
	reports := []piperutils.Path{{Name: "supaResult", Target: supa.resultDir}}
	_ = piperutils.PersistReportsAndLinks("sapSUPAExecuteTests", workdir, utils, reports, nil)

	var err = archiveTestResults(&config, supa.resultDir)
	if err != nil {
		log.Entry().Info("Error: sapSUPAExecuteTests archiveTestResults failed: " + err.Error())
	}

	err = extractReportDataFile(supa.resultDir)
	if errTestFailed != nil {
		log.Entry().WithError(err).Fatal("sapSUPAExecuteTests execution failed")
		return errTestFailed
	}
	if err != nil {
		log.Entry().Info("Error: sapSUPAExecuteTests extractReportDataFile failed: " + err.Error())
	}

	return nil
}

// function takes an input string (sample format: "ABC=abc DEF=def  XYZ=$xyz  ") and returns a string without not-required spaces and resolves all env parameters (given by leading '$')
func getEnvVar(thisEnv string) string {
	_env := strings.TrimSpace(thisEnv)
	// split all substrings by space
	_parr := strings.Split(_env, " ")
	var elem string
	var cmd []string
	var val string
	for i := 0; i < len(_parr); i++ {
		elem = _parr[i]
		elem = strings.TrimSpace(elem)
		// for each substring split by equal sign
		// _sarr := strings.Split(elem, "=")
		// for each substring split by first "="
		idx := strings.Index(elem, "=")
		if idx >= 1 {
			v0 := elem[:idx]
			idx++
			v1 := elem[idx:]
			v := v1
			// for each substring check if leading '$' exists and resolve the variable if necessary
			if strings.HasPrefix(v1, "$") {
				v = os.Getenv(v1[1:])
			}
			val = v0 + "=" + v
			cmd = append(cmd, val)
			//			log.Entry().Debugf("if env variable: " + val)
		} else {
			// only add substring <> ""
			if len(elem) > 0 {
				cmd = append(cmd, elem)
			}
			//			log.Entry().Debugf("else env variable: " + elem)
		}
	}
	if len(cmd) > 0 {
		val = strings.Join(cmd, " ")
	} else {
		val = _env
	}
	log.Entry().Debugf("val: %s", val)
	return val
}

func getDockerEnvParams(config *sapSUPAExecuteTestsOptions, thisEnvs []string) []string {
	cmd := append([]string(nil), thisEnvs...)
	if config.EnvVars != nil {
		s := ""
		leVars := len(config.EnvVars)
		for i := 0; i < leVars; i++ {
			s = config.EnvVars[i]
			s = getEnvVar(s)
			cmd = append(cmd, s)
		}
	}
	if config.SupaKeystoreKey != "" {
		cmd = append(cmd, "SUPA_JCEKS_KEY="+config.SupaKeystoreKey)
	}
	cmd = append(cmd, "AUTOMATE_ENV=Docker_Piper")
	return cmd
}

func getGitParams(config *sapSUPAExecuteTestsOptions, thisEnvs []string, testType int) ([]string, error) {
	cmd := append([]string(nil), thisEnvs...)
	isGitToken := false
	if config.TestRepositoryName != "" {
		cmd = append(cmd, "REPO_NAME="+config.TestRepositoryName)
	}
	if config.TestScriptFolder != "" {
		if testType == TYPE_ALL {
			cmd = append(cmd, "REPO_BASE_PATH="+config.TestScriptFolder)
		} else {
			cmd = append(cmd, "REPO_PATH="+config.TestScriptFolder)
		}
	}
	if config.TestRepositoryBranch != "" {
		cmd = append(cmd, "REPO_BRANCH="+config.TestRepositoryBranch)
	}
	if config.GithubToken != "" {
		cmd = append(cmd, "GITHUB_TOKEN="+config.GithubToken)
		isGitToken = true
	} else {
		log.Entry().Warning("PerformanceSingleUserTest info: GithubToken not defined")
	}
	if config.TestRepository != "" {
		if testType == TYPE_ALL {
			cmd = append(cmd, "REPO_TESTS="+config.TestRepository)
		} else {
			cmd = append(cmd, "REPOSITORY="+config.TestRepository)
		}
		if !isGitToken {
			err := errors.New("PerformanceSingleUserTest error: repository and token not defined")
			return cmd, err
		}
	} else {
		log.Entry().Warning("PerformanceSingleUserTest info: repository not defined")
	}
	return cmd, nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}
	return false
}

func runAllPerformanceSingleUserTest(config *sapSUPAExecuteTestsOptions, _ *telemetry.CustomData, utils sapSUPAExecuteTestsUtils) error {

	log.Entry().Info("This executes the SUPA single user test step with multiple runs of all kind of automation.")
	var err error
	envs := []string{""}
	envs, err = getGitParams(config, envs, TYPE_ALL)
	if err != nil {
		return errors.Wrapf(err, "valid 'all' testcase not defined")
	}
	if len(envs) == 0 {
		log.Entry().Warning(REPO_NOT_FOUND_MESSAGE)
	}
	envs = getDockerEnvParams(config, envs)
	//  log.Entry().Info(ENV_STR, envs)
	utils.SetEnv(envs)
	runTestCmd := RUN_START_ALL
	runCommand := strings.Split(runTestCmd, " ")
	log.Entry().Info(RUN_CMD_STR, runCommand)
	if err = utils.RunExecutable(runCommand[0], runCommand[1:]...); err != nil {
		return errors.Wrapf(err, "failed to execute run All test case: %s", runTestCmd)
	}
	return nil
}

func runSeleniumPerformanceSingleUserTest(config *sapSUPAExecuteTestsOptions, _ *telemetry.CustomData, utils sapSUPAExecuteTestsUtils) error {

	log.Entry().Info("This executes the SUPA single user test step with Selenium automation.")
	var err error
	envs := []string{""}
	envs, err = getGitParams(config, envs, TYPE_DEF)
	if err != nil {
		return errors.Wrapf(err, "valid Selenium testcase not defined")
	}
	if len(envs) == 0 {
		log.Entry().Warning(REPO_NOT_FOUND_MESSAGE)
	}
	if config.TestScript != "" {
		envs = append(envs, "TEST_CASE="+config.TestScript)
	}
	envs = getDockerEnvParams(config, envs)
	// 	log.Entry().Info(ENV_STR, envs)
	utils.SetEnv(envs)
	runTestCmd := RUN_START
	runCommand := strings.Split(runTestCmd, " ")
	log.Entry().Info(RUN_CMD_STR, runTestCmd)
	if err = utils.RunExecutable(runCommand[0], runCommand[1:]...); err != nil {
		return errors.Wrapf(err, "failed to execute run Selenium test: %s", runTestCmd)
	}
	return nil
}

func runWdioPerformanceSingleUserTest(config *sapSUPAExecuteTestsOptions, _ *telemetry.CustomData, utils sapSUPAExecuteTestsUtils) error {

	log.Entry().Info("This executes the SUPA single user test step with Wdio automation.")
	var err error
	envs := []string{""}
	envs, err = getGitParams(config, envs, TYPE_DEF)
	if err != nil {
		return errors.Wrapf(err, "valid Wdio testcase not defined")
	}
	if len(envs) == 0 {
		log.Entry().Warning(REPO_NOT_FOUND_MESSAGE)
	}
	// log.Entry().Info(ENV_STR, envs)
	envs = append(envs, "WDIO_TEST=true")
	if config.TestWdioParams != "" {
		envs = append(envs, "WDIO_PARAMS="+config.TestWdioParams)
	}
	if config.SupaConfig != "" {
		envs = append(envs, PARAM_SUPA_CFG+"="+config.SupaConfig)
	}
	envs = getDockerEnvParams(config, envs)
	if err = runJS5PerformanceSingleUserTest(envs, utils); err != nil {
		return errors.Wrapf(err, "failed to execute run Wdio test: ")
	}
	return nil
}

func runQmatePerformanceSingleUserTest(config *sapSUPAExecuteTestsOptions, _ *telemetry.CustomData, utils sapSUPAExecuteTestsUtils) error {

	log.Entry().Info("This executes the SUPA single user test step with Qmate automation.")
	var err error
	envs := []string{""}
	envs, err = getGitParams(config, envs, TYPE_DEF)
	if err != nil {
		return errors.Wrapf(err, "valid Qmate testcase not defined")
	}
	if len(envs) == 0 {
		log.Entry().Warning(REPO_NOT_FOUND_MESSAGE)
	}
	// log.Entry().Info(ENV_STR, envs)
	envs = append(envs, "QMATE_TEST=true")
	if config.TestQmateParams != "" {
		envs = append(envs, "QMATE_PARAMS="+config.TestQmateParams)
	}
	if config.TestQmateCfg != "" {
		envs = append(envs, "QMATE_CFG="+config.TestQmateCfg)
	}
	if config.SupaConfig != "" {
		envs = append(envs, PARAM_SUPA_CFG+"="+config.SupaConfig)
	}
	envs = getDockerEnvParams(config, envs)
	if err = runJS5PerformanceSingleUserTest(envs, utils); err != nil {
		return errors.Wrapf(err, "failed to execute run Qmate test: ")
	}
	return nil
}

func runNpmPerformanceSingleUserTest(config *sapSUPAExecuteTestsOptions, _ *telemetry.CustomData, utils sapSUPAExecuteTestsUtils) error {

	log.Entry().Info("This executes the SUPA single user test step with NPM automation.")
	var err error
	envs := []string{""}
	envs, err = getGitParams(config, envs, TYPE_DEF)
	if err != nil {
		return errors.Wrapf(err, "valid NPM testcase not defined")
	}
	if len(envs) == 0 {
		log.Entry().Warning(REPO_NOT_FOUND_MESSAGE)
	}
	envs = append(envs, "JS_TEST=npm")
	if config.TestNpmParams != "" {
		envs = append(envs, "NPM_PARAMS="+config.TestNpmParams)
	}
	if config.SupaConfig != "" {
		envs = append(envs, PARAM_SUPA_CFG+"="+config.SupaConfig)
	}
	envs = getDockerEnvParams(config, envs)
	if err = runJS5PerformanceSingleUserTest(envs, utils); err != nil {
		return errors.Wrapf(err, "failed to execute run NPM test: ")
	}
	return nil
}

func runUiveri5PerformanceSingleUserTest(config *sapSUPAExecuteTestsOptions, _ *telemetry.CustomData, utils sapSUPAExecuteTestsUtils) error {

	log.Entry().Info("This executes the SUPA single user test step with UiVeri5 automation.")
	var err error
	envs := []string{""}
	envs, err = getGitParams(config, envs, TYPE_DEF)
	if err != nil {
		return errors.Wrapf(err, "valid UiVeri5 testcase not defined")
	}
	if len(envs) == 0 {
		log.Entry().Warning(REPO_NOT_FOUND_MESSAGE)
	}
	envs = append(envs, "UI5_TEST=true")
	if config.TestUiVeri5Params != "" {
		envs = append(envs, "UI5_PARAMS="+config.TestUiVeri5Params)
	}
	if config.SupaConfig != "" {
		envs = append(envs, PARAM_SUPA_CFG+"="+config.SupaConfig)
	}
	envs = getDockerEnvParams(config, envs)

	if err = runJS5PerformanceSingleUserTest(envs, utils); err != nil {
		return errors.Wrapf(err, "failed to execute run UiVeri5 test: ")
	}
	return nil
}

func runJS5PerformanceSingleUserTest(thisEnvs []string, utils sapSUPAExecuteTestsUtils) error {
	//	log.Entry().Info(ENV_STR, thisEnvs)
	utils.SetEnv(thisEnvs)
	runTestCmd := RUN_START_JS
	runCommand := strings.Split(runTestCmd, " ")
	log.Entry().Info(RUN_CMD_STR, runCommand)
	if err := utils.RunExecutable(runCommand[0], runCommand[1:]...); err != nil {
		err := errors.New("failed to execute run JS test case: " + runTestCmd)
		return err
	}
	return nil
}

func runJMeterPerformanceSingleUserTest(config *sapSUPAExecuteTestsOptions, _ *telemetry.CustomData, utils sapSUPAExecuteTestsUtils) error {

	log.Entry().Info("This executes the SUPA single user test step with JMeter automation.")
	var err error
	envs := []string{""}
	envs, err = getGitParams(config, envs, TYPE_DEF)
	if err != nil {
		return errors.Wrapf(err, "valid JMeter testcase not defined")
	}
	if len(envs) == 0 {
		log.Entry().Warning(REPO_NOT_FOUND_MESSAGE)
	}
	if config.TestJMeterParams != "" {
		envs = append(envs, "JMETER_PARAMS="+config.TestJMeterParams)
	}
	if config.TestJMeterScript != "" {
		envs = append(envs, "JMETER_SCRIPT="+config.TestJMeterScript)
	}
	if config.SupaConfig != "" {
		envs = append(envs, PARAM_SUPA_CFG+"="+config.SupaConfig)
	}
	envs = getDockerEnvParams(config, envs)

	//	log.Entry().Info(ENV_STR, envs)
	utils.SetEnv(envs)
	runTestCmd := RUN_START
	runCommand := strings.Split(runTestCmd, " ")
	log.Entry().Info(RUN_CMD_STR, runCommand)
	if err = utils.RunExecutable(runCommand[0], runCommand[1:]...); err != nil {
		return errors.Wrapf(err, "failed to execute run JMeter test: %s", runTestCmd)
	}
	return nil
}

func runKryptonPerformanceSingleUserTest(config *sapSUPAExecuteTestsOptions, _ *telemetry.CustomData, utils sapSUPAExecuteTestsUtils) error {

	log.Entry().Info("This executes the SUPA single user test step with Krypton automation.")
	var err error
	envs := []string{""}
	envs, err = getGitParams(config, envs, TYPE_DEF)
	if err != nil {
		return errors.Wrapf(err, "valid Krypton testcase not defined")
	}
	if len(envs) == 0 {
		log.Entry().Warning(REPO_NOT_FOUND_MESSAGE)
	}
	envs = append(envs, "KRYPTON_TEST=True")
	if config.TestKryptonParams != "" {
		log.Entry().Info("This test executes a Krypton test in SUPA Docker container: " + config.TestKryptonParams)
		envs = append(envs, "KRYPTON_PARAMS="+config.TestKryptonParams)
	}
	if config.TestKryptonScript != "" {
		envs = append(envs, "KRYPTON_SCRIPT="+config.TestKryptonScript)
	}
	if config.TestKryptonRuns != "" {
		envs = append(envs, "KRYPTON_NR="+config.TestKryptonRuns)
	}
	envs = getDockerEnvParams(config, envs)

	//	log.Entry().Info(ENV_STR, envs)
	utils.SetEnv(envs)
	runTestCmd := RUN_START
	runCommand := strings.Split(runTestCmd, " ")
	log.Entry().Info(RUN_CMD_STR, runCommand)
	if err = utils.RunExecutable(runCommand[0], runCommand[1:]...); err != nil {
		return errors.Wrapf(err, "failed to execute run Krypton test: %s", runTestCmd)
	}
	return nil
}

func runAnyTestInContainer(config *sapSUPAExecuteTestsOptions, _ *telemetry.CustomData, utils sapSUPAExecuteTestsUtils) error {

	log.Entry().Info("This executes any command in SUPA Docker container.")
	var err error
	envs := []string{""}
	if config.CmdExe != "" {
		envs = append(envs, "command="+config.CmdExe)
	} else {
		err := errors.New("missing CMD parameters for any test case")
		return errors.Wrapf(err, "valid Any testcase not defined")
	}
	envs = getDockerEnvParams(config, envs)

	log.Entry().Info(ENV_STR, envs)
	utils.SetEnv(envs)
	runTestCmd := config.CmdExe + " " + config.CmdParams
	runCommand := strings.Split(runTestCmd, " ")
	log.Entry().Info(RUN_CMD_STR, runCommand)
	if err = utils.RunExecutable(runCommand[0], runCommand[1:]...); err != nil {
		return errors.Wrapf(err, "failed to execute run Any test: %s", runTestCmd)
	}
	return nil
}

func archiveTestResults(config *sapSUPAExecuteTestsOptions, resultDir string) error {

	if config.ArchiveSUPAResult {
		var path string
		if _, err := os.Stat(resultDir); os.IsNotExist(err) {
			return errors.Wrapf(err, "Error archiveTestResults - not exist resultDir")
		}
		// use the parent folder from resultDir: supaData/results
		useDir := filepath.Join(resultDir, "..")
		//		dirs := strings.Split(resultDir, os.PathSeparator)
		//		useDir := dirs[len(dirs)-1]
		log.Entry().Info("archiveTestResults - resultDir: ", useDir)
		targetPath, err := os.Getwd()
		log.Entry().Info("archiveTestResults - targetPath: ", targetPath)
		path = filepath.Join(targetPath, SUPA_RESULT_PACKAGE)
		archive, err := os.Create(filepath.Clean(path))
		if err != nil {
			return errors.Wrapf(err, "Error archiveTestResults - create file: %s", path)
		}
		defer archive.Close()
		w := zip.NewWriter(archive)
		// add files from folder useDir to zip package
		addFiles(w, useDir, "" /* targetPath */)

		err = w.Close()
		if err != nil {
			return errors.Wrapf(err, "Error archiveTestResults - close zip file: %s", path)
		}
	}
	return nil
}

func addFiles(w *zip.Writer, basePath, baseZipPath string) {
	// Open the Directory
	files, err := ioutil.ReadDir(basePath)
	if err != nil {
		log.Entry().Info("Error read the dir basePath: ", basePath)
		return
	}

	for _, file := range files {
		f1 := filepath.Join(basePath, file.Name())
		log.Entry().Info("archiveTestResults - read and add the file: ", basePath, " - ", file.Name())
		if !file.IsDir() {
			dat, err := ioutil.ReadFile(filepath.Clean(f1))
			if err != nil {
				log.Entry().Info("Error read the file: ", f1)
				continue // continue next file
			}
			// Add some files to the archive.
			f2 := filepath.Join(baseZipPath, file.Name())
			f3, err := w.Create(f2)
			if err != nil {
				log.Entry().Info("Error create the file: ", f2)
				continue // continue next file
			}
			_, err = f3.Write(dat)
			if err != nil {
				log.Entry().Info("Error write the file: ", f3)
				continue // continue next file
			}
		} else if file.IsDir() {
			// Recursive walk through folders
			nextBase := filepath.Join(basePath, file.Name())
			nextBaseZipPath := filepath.Join(baseZipPath, file.Name())
			msg := "Recursing and Adding SubDir and file: " + nextBase + " - " + file.Name() + " - " + nextBaseZipPath
			log.Entry().Info(msg)
			addFiles(w, nextBase, nextBaseZipPath)
		}
	}
}

func extractReportDataFile(resultDir string) error {
	files, err := ioutil.ReadDir(resultDir)
	if err != nil {
		return errors.Wrapf(err, "failed to extractReportDataFile, read dir: %s", resultDir)
	}
	// descending sort by date
	if len(files) > 1 {
		sort.Slice(files, func(i, j int) bool {
			fileI := files[i]
			fileJ := files[j]
			return fileI.ModTime().After(fileJ.ModTime())
		})
	}
	var path string
	for _, f := range files {
		// This returns an *os.FileInfo type
		path = filepath.Join(resultDir, f.Name())
		file, err := os.Open(filepath.Clean(path))
		if err != nil {
			log.Entry().Info("extractReportDataFile - open file: ", path)
			continue // continue next file
		}
		defer file.Close()
		fileInfo, err := file.Stat()
		if err != nil {
			log.Entry().Info("extractReportDataFile - stat file: ", path)
			continue // continue next file
		}
		if fileInfo.IsDir() {
			log.Entry().Info("extractReportDataFile - file: ", f.Name())
			targetPath, err1 := os.Getwd()
			if err1 != nil {
				msg := "failed to extractReportDataFile: os.pwd not found"
				log.Entry().Warning(msg)
				continue // continue next file
			}
			// path is the path to SUPA result zip file f
			// extract the cumulus data file supa_upload_cumulus.json and copy as data.json to the working path
			err1 = extractFile(path, f.Name(), targetPath)
			if err1 != nil {
				msg := "failed to extractReportDataFile: extract " + path + " - error: " + err1.Error()
				log.Entry().Warning(msg)
				continue // continue next file
			}
		}
	}
	return nil
}

// path is the path to SUPA result zip file f
// extract the cumulus data file supa_upload_cumulus.json and copy as data.json to the working path
func extractFile(path string, file string, targetPath string) error {
	log.Entry().Info("extractFile - path: ", path)
	file = file + ".zip"
	// concatenate the path with data zip file name
	filePath := filepath.Join(path, file)
	archive, err := zip.OpenReader(filePath)
	if err != nil {
		msg := "extractFile - failed to open zip file: " + file
		err = errors.New(msg)
		return err
	}
	defer archive.Close()
	log.Entry().Info("extractFile - unzip filepath: ", filePath)
	// supa_upload_cumulus.json
	for _, f := range archive.File {
		if strings.Contains(f.Name, CUMULUS_SOURCE_FILE) {
			//			log.Entry().Info("extractFile - unzip file: ", f.Name)
			filePath = path + string(os.PathSeparator) + f.Name
			// open and read the cumulus data fril from archive
			fileInArchive, err1 := f.Open()
			if err1 != nil {
				err = errors.New("extractFile - failed to extractFile fileInArchive")
			}
			defer fileInArchive.Close()
			//			log.Entry().Info("extractFile - unzip filePath: ", filePath)
			// create the target file to write

			targetFile := filepath.Clean(filepath.Join(targetPath, CUMULUS_DATA_FILE))
			dstFile, err1 := os.OpenFile(targetFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err1 != nil {
				err = errors.New("extractFile - failed to create target dstFile")
			}
			defer dstFile.Close()
			//			log.Entry().Info("extractFile - fileInArchive: ", fileInArchive)
			// copy the data file to the target file
			limitOnDecompressedBytes := int64(500000) // 500KB, added file size limit
			limitedReader := io.LimitReader(fileInArchive, limitOnDecompressedBytes)
			_, err1 = io.Copy(dstFile, limitedReader)
			if err1 != nil {
				err = errors.New("extractFile - failed to Copy")
			}
			log.Entry().Info("extractFile - dest File: ", dstFile.Name())
			log.Entry().Info("extractFile - source file: ", f.Name)
			break
		}
	}

	if err != nil {
		return errors.Wrapf(err, "failed to extractFile: %s", path)
	}
	return nil
}
