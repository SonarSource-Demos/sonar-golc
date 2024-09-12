package getter

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/briandowns/spinner"
	getter "github.com/hashicorp/go-getter"
)

func extractLastString(url string) string {
	/*parts := strings.Split(url, "/")
	return parts[len(parts)-1]*/

	return filepath.Base(url)
}

func Getter(src string) (string, error) {
	RepoString := extractLastString(src)

	spinner := newSpinner(fmt.Sprintf("\r Extracting files from %s \n", RepoString))
	spinner.Color("green", "bold")
	messageF := ""
	spinner.FinalMSG = messageF
	spinner.Start()
	defer spinner.Stop()

	suffix, err := randomSuffix()
	if err != nil {
		return "", err
	}

	dst := filepath.Join(os.TempDir(), fmt.Sprintf("gcloc-extract-%s", suffix))
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	client := &getter.Client{
		Src: src,
		Dst: dst,
		Pwd: pwd,
		//Mode: getter.ClientModeAny,
		Mode: getter.ClientModeDir,
	}

	if err := client.Get(); err != nil {
		return "", err
	}

	symLink, err := isSymLink(dst)
	if err != nil {
		return "", err
	}

	if symLink {
		origin, err := os.Readlink(dst)
		if err != nil {
			return "", err
		}

		return origin, nil
	}

	return dst, nil
}

func newSpinner(text string) *spinner.Spinner {
	return spinner.New(
		spinner.CharSets[35],
		100*time.Millisecond,
		spinner.WithSuffix(text),
	)
}

func randomSuffix() (string, error) {
	randBytes := make([]byte, 16)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(randBytes), nil
}

func isSymLink(path string) (bool, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return false, err
	}

	return info.Mode()&os.ModeSymlink != 0, nil
}
