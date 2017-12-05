package exec

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/satori/go.uuid"
)

// Context is a set of information surrounding how to execute a certain executable. It provides utilities for
// executing the same executable multiple times with the same working directory and log path, and allows for
// adding files to the working directory, setting the environment, and retrieving log files.
type Context struct {
	executablePath string
	WorkingDir     *string
	LogPath        *string
	envVars        map[string]string
}

type Result struct {
	Pid  int
	Wait func() error
}

func NewContext(executablePath string, logPath, workingDir *string, envVars map[string]string) (*Context, error) {
	if workingDir != nil {
		err := os.MkdirAll(*workingDir, 0770)
		if err != nil {
			return nil, err
		}
	}

	if logPath != nil {
		err := os.MkdirAll(*logPath, 0770)
		if err != nil {
			return nil, err
		}
	}

	ec := &Context{
		executablePath: executablePath,
		WorkingDir:     workingDir,
		LogPath:        logPath,
		envVars:        envVars,
	}

	return ec, nil
}

// SetEnvVars overrides the existing EnvVars for the
func (ec *Context) SetEnvVars(envVars map[string]string) {
	ec.envVars = envVars
}

// AddFile adds a new file with the contents to the working directory of the Context.
func (ec *Context) AddFile(filename string, content []byte) error {
	if ec.WorkingDir == nil {
		return fmt.Errorf("cannot add file to ExecContext with no WorkingDir")
	}

	filePath := filepath.Join(*ec.WorkingDir, filename)
	return ioutil.WriteFile(filePath, content, 0660)
}

// Exec begins executing the executable, and returns a Result containing the Pid of the process, as well as
// a function that will wait for the process to complete and clean up, returning an error if the process failed.
// Exec will forward stdout and stderr of the spawned process to the respective io.Writers passed in. cleanup
// should do any extra cleanup necessary required by the calling function (such as closing a log file that
// stdout/stderr are being piped to). If there is no cleanup to be done, nil may be passed.
func (ec *Context) Exec(stdout, stderr io.Writer, cleanup func(), args ...string) (*Result, error) {
	cmd := exec.Command(ec.executablePath, args...)

	// Set up EnvVars
	for k, v := range ec.envVars {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Set working dir
	if ec.WorkingDir != nil {
		cmd.Dir = *ec.WorkingDir
	}

	cmd.Stdout = stdout
	cmd.Stderr = stderr

	// Start process
	err := cmd.Start()
	if err != nil {
		cleanup()
		return nil, err
	}

	// Create function that waits for process to exit and cleans up
	waitFunc := func() error {
		if cleanup != nil {
			defer cleanup()
		}

		return cmd.Wait()
	}

	result := &Result{
		Pid:  cmd.Process.Pid,
		Wait: waitFunc,
	}
	return result, nil
}

// ExecWithLogFile runs Exec, piping stdout and stderr to a log file which is created in the LogPath
// of the exec context, whose filename will be logPrefix-<generated-UUID>.log. ExecWithLogFile returns
// the Result from Exec, along with the log's filename.
func (ec *Context) ExecWithLogFile(logPrefix string, args ...string) (*Result, string, error) {
	// Create logfile and redirect stdout/stderr to it
	logFilename := fmt.Sprintf("%s-%s.log", logPrefix, uuid.NewV4().String())
	logFilePath, err := ec.getLogfilePath(logFilename)
	if err != nil {
		return nil, "", err
	}

	logFile, err := os.Create(logFilePath)
	if err != nil {
		return nil, "", err
	}

	cleanup := func() {
		logFile.Close()
	}

	result, err := ec.Exec(logFile, logFile, cleanup, args...)
	return result, logFilename, err
}

// ExecWithLogFileSync runs ExecWithLogFile but blocks until the process has completed, returning the
// log filename.
func (ec *Context) ExecWithLogFileSync(logPrefix string, args ...string) (string, error) {
	result, logFilename, err := ec.ExecWithLogFile(logPrefix, args...)
	if err != nil {
		return logFilename, err
	}

	err = result.Wait()
	return logFilename, err
}

// LogFile returns an io.ReadCloser for the log file with the given logFilename in the LogPath.
func (ec *Context) LogFile(logFilename string) (io.ReadCloser, error) {
	logFilePath, err := ec.getLogfilePath(logFilename)
	if err != nil {
		return nil, err
	}

	return os.Open(logFilePath)
}

// ExecSync blocks while executing, and returns stdout and stderr.
func (ec *Context) ExecSync(args ...string) (string, string, error) {
	var stdoutBuf bytes.Buffer
	var stderrBuf bytes.Buffer
	stdoutWriter := bufio.NewWriter(&stdoutBuf)
	stderrWriter := bufio.NewWriter(&stderrBuf)

	result, err := ec.Exec(stdoutWriter, stderrWriter, nil, args...)
	if err != nil {
		return "", "", nil
	}

	err = result.Wait()

	// FIXME: handle errors from flush
	stdoutWriter.Flush()
	stderrWriter.Flush()

	return stdoutBuf.String(), stderrBuf.String(), err
}

func (ec *Context) getLogfilePath(logFilename string) (string, error) {
	if ec.LogPath == nil {
		return "", fmt.Errorf("cannot get logfile path because LogPath is nil")
	}
	return filepath.Join(*ec.LogPath, logFilename), nil
}
