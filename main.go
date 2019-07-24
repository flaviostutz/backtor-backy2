package main

import (
	"flag"
	"fmt"
	"regexp"

	conductor "github.com/ggrcha/conductor-go-client"
	"github.com/ggrcha/conductor-go-client/task"

	"github.com/sirupsen/logrus"
)

var (
	sourcePath     string
	repoDir        string
	resticPassword string
)

func main() {
	logLevel := flag.String("log-level", "debug", "debug, info, warning, error")
	conductorURL0 := flag.String("conductor-url", "", "Conductor API URL")
	sourcePath0 := flag.String("source-path", "/backup-source", "Backup source path")
	repoDir0 := flag.String("repo-dir", "/backup-repo", "Restic repository of backups")
	resticPassword0 := flag.String("restic-password", "", "Restic repository password")
	flag.Parse()

	switch *logLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
		break
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
		break
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
		break
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	sourcePath = *sourcePath0
	repoDir = *repoDir0
	resticPassword = *resticPassword0

	if sourcePath == "" {
		logrus.Errorf("'--source-path' is required")
		panic(1)
	}
	if repoDir == "" {
		logrus.Errorf("'--repo-dir' is required")
		panic(1)
	}
	if resticPassword == "" {
		logrus.Errorf("'--restic-password' is required")
		panic(1)
	}
	if *conductorURL0 == "" {
		logrus.Errorf("'--conductor-url' is required")
		panic(1)
	}

	logrus.Info("====Starting Restic Conductor Worker====")

	initRepo()

	c := conductor.NewConductorWorker(*conductorURL0, 1, 1000)

	c.Start("backup", backupTask, false)
	c.Start("remove", removeTask, true)
}

func backupTask(t *task.Task) (tr *task.TaskResult, err error) {
	logrus.Debugf("Executing backupTask")

	bn, ok := t.InputData["backupName"]
	if !ok {
		return tr, fmt.Errorf("'backupName' is required as Input data")
	}

	backupName := bn.(string)
	logrus.Debugf("Creating backup. backupName=%s", backupName)

	dataID, dataSizeMB, err := createNewBackup(backupName)
	if err != nil {
		return nil, err
	}

	tr = task.NewTaskResult(t)
	output := map[string]interface{}{
		"dataId":     dataID,
		"dataSizeMB": dataSizeMB,
	}
	tr.OutputData = output
	tr.Status = task.COMPLETED

	return tr, nil
}

func removeTask(t *task.Task) (tr0 *task.TaskResult, err0 error) {
	logrus.Debugf("Executing removeTask")

	bn, ok := t.InputData["backupName"]
	if !ok {
		return tr0, fmt.Errorf("'backupName' is required as Input data")
	}
	backupName := bn.(string)

	di, ok := t.InputData["dataId"]
	if !ok {
		return tr0, fmt.Errorf("'backupName' is required as Input data")
	}
	dataID := di.(string)

	logrus.Debugf("Deleting backup. backupName=%s dataID=%s", backupName, dataID)

	err := deleteBackup(dataID)
	if err != nil {
		return nil, err
	}

	tr := task.NewTaskResult(t)
	output := map[string]interface{}{}
	tr.OutputData = output
	tr.Status = task.COMPLETED

	return tr, nil
}

func initRepo() error {
	logrus.Debugf("Checking if Restic repo %s was already initialized", repoDir)
	result, err := ExecShellf("restic snapshots -r %s", repoDir)
	if err != nil {
		logrus.Debugf("Couldn't access Restic repo. Trying to create it. err=%s", err)
		_, err := ExecShellf("restic init -r %s", repoDir)
		if err != nil {
			logrus.Debugf("Error creating Restic repo: %s %s", err, result)
			return err
		}
		logrus.Infof("Restic repo created successfuly")
	} else {
		logrus.Infof("Restic repo already exists and is accessible")
	}
	return nil
}

func createNewBackup(backupName string) (dataID0 string, dataSizeMB0 int, err0 error) {
	logrus.Infof("createNewBackup() backupName=%s", backupName)

	logrus.Infof("Calling Restic...")
	result, err := ExecShellf("restic backup /backup-source/%s -r %s", backupName, repoDir)
	if err != nil {
		return "", -1, err
	}
	logrus.Debugf("result: %s", result)
	rex, _ := regexp.Compile("snapshot ([0-9a-zA-z]+) saved")
	id := rex.FindStringSubmatch(result)
	success := (len(id) == 2)
	if !success {
		logrus.Warnf("Snapshot not created. result=%s", result)
	}

	dataID := id[1]
	logrus.Infof("Backup finished")

	dataSizeMB := 111

	return dataID, dataSizeMB, nil
}

func deleteBackup(dataID string) error {
	logrus.Debugf("deleteBackup dataID=%s", dataID)

	logrus.Debugf("Backup dataID=%s found. Proceeding to deletion", dataID)
	result, err := ExecShellf("restic forget %s -r %s", dataID, repoDir)
	if err != nil {
		return err
	}
	logrus.Debugf("result: %s", result)

	rex, _ := regexp.Compile("removed snapshot ([0-9a-zA-z]+)")
	id := rex.FindStringSubmatch(result)
	if len(id) != 2 {
		return fmt.Errorf("Couldn't find returned id from response")
	}
	if id[1] != dataID {
		return fmt.Errorf("Returned id from forget is different from requested. %s != %s", id[1], dataID)
	}

	logrus.Debugf("Delete dataID %s successful", dataID)
	return nil
}
