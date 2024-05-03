package utils

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func fileExists(fileexclusion string) bool {
	_, err := os.Stat(fileexclusion)
	return !os.IsNotExist(err)
}

func isFileEmpty(fileexclusion string) (bool, error) {
	fileInfo, err := os.Stat(fileexclusion)
	if err != nil {
		return false, err
	}

	return fileInfo.Size() == 0, nil
}

func searchStringInFile(fileexclusion string, target string) (bool, error) {
	file, err := os.Open(fileexclusion)
	if err != nil {
		return false, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, target) {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

func CheckCLOCignoreFile(fileexclusion string, target string) (bool, error) {

	if fileExists(fileexclusion) {
		isEmpty, err := isFileEmpty(fileexclusion)
		if err != nil {
			fmt.Println("❌ -- Stack: utils.checkCLOCignoreFile Empty Test .clocignore -- ", err)
			return false, err
		}

		if isEmpty {
			//The file exists but is empty
			return false, nil
		} else {
			//The file exists but is not empty
			found, err := searchStringInFile(fileexclusion, target)
			if err != nil {
				fmt.Println("❌ -- Stack: utils.checkCLOCignoreFile search exclusion -- ", err)
				return false, err
			}

			if found {
				//Repo was found in the file
				return true, nil
			} else {
				return false, nil
			}
		}
	} else {
		//The file does not exist
		return false, nil
	}
}
