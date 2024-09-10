package gogitv2

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/SonarSource-Demos/sonar-golc/pkg/utils"
	"github.com/go-git/go-git"
	"github.com/go-git/go-git/plumbing/protocol/packp/capability"
	"github.com/go-git/go-git/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing"
)

func Getrepos(src, branch, token string) (string, error) {

	loggers := utils.NewLogger()
	suffix, err := randomSuffix()
	if err != nil {
		return "", err
	}

	dst := filepath.Join(os.TempDir(), fmt.Sprintf("gcloc-extract-%s", suffix))
	//pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	log.SetOutput(os.Stderr)

	url := fmt.Sprintf("https://github.com/%s/%s/archive/refs/heads/%s.zip", owner, repo, branch)

	transport.UnsupportedCapabilities = []capability.Capability{
		capability.ThinPack,
	}

	_, err = git.PlainClone(dst, false, &git.CloneOptions{
		URL: src,

		ReferenceName: plumbing.NewBranchReferenceName(branch),
		//ReferenceName: plumbing.ReferenceName(branch),

		SingleBranch: true,
		Depth:        1,
	})

	if err != nil {
		re := regexp.MustCompile(`(https?:\/\/)[^@]+(@)`)
		maskedSrc := re.ReplaceAllString(src, "${1}*****${2}")
		//fmt.Printf("\n--❌ Stack: gogit.Getrepos Git Branch %s - %s-- Source: %s -", plumbing.Main, err, maskedSrc)
		loggers.Errorf("\r\t\t\t\t❌ Stack: gogit.Getrepos Git Branch %s - %s-- Source: %s -", plumbing.Main, err, maskedSrc)

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
