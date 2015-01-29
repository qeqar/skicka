package main

import (
	"fmt"
	"github.com/google/skicka/gdrive"
)

type removeDirectoryError struct {
	path        string
	invokingCmd string
}

func (err removeDirectoryError) Error() string {
	msg := ""
	if err.invokingCmd != "" {
		msg += fmt.Sprintf("%s: ", err.invokingCmd)
	}
	return fmt.Sprintf("%s%s: is a directory", msg, err.path)
}

var rmSyntaxError CommandSyntaxError = CommandSyntaxError{
	Cmd: "rm",
	Msg: "drive path cannot be empty.\n" +
		"Usage: rm [-r, -s] drive path",
}

func Rm(args []string) {
	recursive, skipTrash := false, false
	var drivePath string

	for _, arg := range args {
		switch {
		case arg == "-r":
			recursive = true
		case arg == "-s":
			skipTrash = true
		case drivePath == "":
			drivePath = arg
		default:
			printErrorAndExit(rmSyntaxError)
		}
	}

	if drivePath == "" {
		printErrorAndExit(rmSyntaxError)
	}

	if err := checkRmPossible(drivePath, recursive); err != nil {
		if _, ok := err.(gdrive.FileNotFoundError); ok {
			// if there's an encrypted version on drive, let the user know and exit
			oldPath := drivePath
			drivePath += encryptionSuffix
			if err := checkRmPossible(drivePath, recursive); err == nil {
				printErrorAndExit(fmt.Errorf("skicka rm: Found no file with path %s, but found encrypted version with path %s.\n"+
					"If you would like to rm the encrypted version, re-run the command adding the %s extension onto the path.",
					oldPath, drivePath, encryptionSuffix))
			}
		}
		printErrorAndExit(err)
	}

	f, err := gd.GetFile(drivePath)
	if err != nil {
		printErrorAndExit(err)
	}

	if skipTrash {
		err = gd.DeleteFile(f)
	} else {
		err = gd.TrashFile(f)
	}
	if err != nil {
		printErrorAndExit(err)
	}
}

func checkRmPossible(path string, recursive bool) error {
	invokingCmd := "skicka rm"

	driveFile, err := gd.GetFile(path)
	if err != nil {
		switch err.(type) {
		case gdrive.FileNotFoundError:
			return gdrive.NewFileNotFoundError(path, invokingCmd)
		default:
			return err
		}
	}

	if !recursive && gdrive.IsFolder(driveFile) {
		return removeDirectoryError{
			path:        path,
			invokingCmd: invokingCmd,
		}
	}

	return nil
}