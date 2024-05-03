package gogit

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func Getrepos(src, branch string) (string, error) {

	suffix, err := randomSuffix()
	if err != nil {
		return "", err
	}

	dst := filepath.Join(os.TempDir(), fmt.Sprintf("gcloc-extract-%s", suffix))
	//pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	_, err = git.PlainClone(dst, false, &git.CloneOptions{
		URL: src,
		// if you want the progress bar
		//Progress:      os.Stdout,
		ReferenceName: plumbing.NewBranchReferenceName(branch),
		SingleBranch:  true,
	})

	if err != nil {
		fmt.Printf("\n--❌ Stack: gogit.Getrepos Git Branch %s %s-- Source: %s", plumbing.NewBranchReferenceName(branch), err, src)
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
