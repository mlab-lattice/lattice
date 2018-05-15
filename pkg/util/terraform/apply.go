package terraform

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
)

func Apply(workDirectory string, config *Config) (string, error) {

	fmt.Fprintln(os.Stderr, "GROOOOOOD")

	tec, err := NewTerrafromExecContext(workDirectory, nil)
	if err != nil {
		return "", err
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return "", err
	}

	prettyprint(configBytes)

	err = tec.AddFile("config.tf.json", configBytes)
	if err != nil {
		return "", err
	}

	result, logfile, err := tec.Init()
	if err != nil {
		return logfile, err
	}

	err = result.Wait()
	if err != nil {
		return logfile, err
	}

	result, logfile, err = tec.Apply(nil)
	if err != nil {
		return logfile, err
	}

	return logfile, result.Wait()
}

func prettyprint(b []byte) {
	fmt.Println("PRETTY PRINT 111")
	fmt.Fprintln(os.Stderr, "PRETTY PRINT 2222")
	var out bytes.Buffer
	err := json.Indent(&out, b, "", "  ")
	if err == nil {
		fmt.Fprintf(os.Stderr, "%s\n", out.Bytes())
	} else {
		fmt.Fprintf(os.Stderr, "PRETTY PRINT ERROR %v\n", err)
	}

}
