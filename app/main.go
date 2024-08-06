package app

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/go-git/go-git/v5"
)


var (
    TARGET string
    IGNORE []string
    GIT_REPOS []gitInfo
)

type gitInfo struct {
    LocalPath   string
    RemoteURLs  []string
    IsClean     bool
    Status      []string
}

// isDirectory determines if a file represented
// by `path` is a directory or not
func isDirectory(path string) bool {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fileInfo.IsDir()
}

// Expects an absolute path to a directory and checks if it contains
// a .git dir indicating that this path is a git repository.
func isGitRepo(dirPath string) (bool, error) {
    entries, err := os.ReadDir(dirPath)
    if err != nil {
        return false, err
    }

    for _, entry := range entries {
        if entry.IsDir() && entry.Name() == ".git" {
            return true, nil 
        }
    }
    return false, nil 
}

// Recursive helper for scanning git repos
func recurseInSearchOfGit(dirPath string) (error) {
    entries, err := os.ReadDir(dirPath)
    if err != nil {
        return err
    }

    path, err := filepath.Abs(dirPath)
    if err != nil {
        return err
    }

    for _, entry := range entries {
        if entry.IsDir() {
            entryPath := filepath.Join(path, entry.Name())
            isGitRepo, err := isGitRepo(entryPath)
            if err != nil {
                return err
            }
            if isGitRepo {
                info, err := collectGitInfo(entryPath)
                if err != nil {
                    return err
                }
                GIT_REPOS = append(GIT_REPOS, info)
            } else {
                if !slices.Contains(IGNORE, dirPath) {
                    recurseInSearchOfGit(entryPath)
                }
            }
        }
    }
    return nil
}

// Build gitInfo from a path to a local repository
func collectGitInfo(path string) (gitInfo, error)  {
    if !isDirectory(path) {
        return gitInfo{}, errors.New("Error collecting git info.  Provided path is not a directory.")
    } else {
        repo, err := git.PlainOpen(path)
        if err != nil {
            return gitInfo{}, err
        }

        worktree, err := repo.Worktree()
        if err != nil {
            return gitInfo{}, err
        }

        status, err := worktree.Status()
        if err != nil {
            return gitInfo{}, err
        }

        info := gitInfo{
        	LocalPath:  path,
        	IsClean:    status.IsClean(),
        	Status:     strings.Split(status.String(), "\n"),
        }
        return info, nil
    }
}

// Print results of this audit to stdout
func printReport() {
    fmt.Println("Audit Report for: " + TARGET)
    fmt.Println()
    numOfDirtyRepos := 0
    for _, repo := range GIT_REPOS {
        var clean string
        if repo.IsClean {
            clean = "CLEAN"
        } else {
            clean = "DIRTY"
            numOfDirtyRepos = numOfDirtyRepos + 1
        }
        fmt.Println("[" + clean + "] " + repo.LocalPath)
        if !repo.IsClean {
            fmt.Println("\tStatus: ")
            for _, file := range repo.Status {
                fmt.Println("\t\t" + file)
            }
        }
        fmt.Println()
    }
    fmt.Println(strconv.Itoa(numOfDirtyRepos) + " dirty repositories")
}

// takes a string [target] and searches for git repos within the specified path.
// Generates and outputs a report of the status of found git repositories.
func Main(target string, ignore []string)  {
    // Validate target
    target = os.Expand(target, os.Getenv)
    if !isDirectory(target) {
        println("Specified target '" + target + "' is not a valid directory")
        return
    }
    TARGET = target

    // Expand ignore list envvars and set to global
    for _, p := range ignore {
        IGNORE = append(IGNORE, os.Expand(p, os.Getenv))
    }

    // Begin Audit
    fmt.Println("Beginning an audit of git repositories in: " + target)

    path, err := filepath.Abs(target)
    if err != nil {
        println(err.Error())
        return
    }
    if !slices.Contains(IGNORE, target) {
        entries, err := os.ReadDir(target)
        if err != nil {
            println(err.Error())
            return
        }
        for _, entry := range entries {
            if entry.IsDir() {
                dirPath := filepath.Join(path, entry.Name())
                isGitRepo, err := isGitRepo(dirPath)
                if err != nil {
                    println(err.Error())
                    return
                }
                if isGitRepo {
                    info, err := collectGitInfo(dirPath)
                    if err != nil {
                        println(err.Error())
                        return
                    }
                    GIT_REPOS = append(GIT_REPOS, info)
                } else {
                    if !slices.Contains(IGNORE, dirPath) {
                        recurseInSearchOfGit(dirPath)
                    }
                }
            }
        }
    }

    printReport()
}

