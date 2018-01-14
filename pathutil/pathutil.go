package pathutil

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Normalize the following forms into "github.com/user/name":
// 1. user/name[.git]
// 2. github.com/user/name[.git]
// 3. [git|http|https]://github.com/user/name[.git]
func NormalizeRepos(rawReposPath string) (ReposPath, error) {
	rawReposPath = filepath.ToSlash(rawReposPath)
	paths := strings.Split(rawReposPath, "/")
	if len(paths) == 3 {
		return ReposPath(strings.TrimSuffix(rawReposPath, ".git")), nil
	}
	if len(paths) == 2 {
		return ReposPath(strings.TrimSuffix("github.com/"+rawReposPath, ".git")), nil
	}
	if paths[0] == "https:" || paths[0] == "http:" || paths[0] == "git:" {
		path := strings.Join(paths[len(paths)-3:], "/")
		return ReposPath(strings.TrimSuffix(path, ".git")), nil
	}
	return ReposPath(""), errors.New("invalid format of repository: " + rawReposPath)
}

type ReposPath string
type ReposPathList []ReposPath

func (path *ReposPath) String() string {
	return string(*path)
}

func (list ReposPathList) Strings() []string {
	// TODO: Use unsafe
	result := make([]string, 0, len(list))
	for i := range list {
		result = append(result, string(list[i]))
	}
	return result
}

func NormalizeLocalRepos(name string) (ReposPath, error) {
	if !strings.Contains(name, "/") {
		return ReposPath("localhost/local/" + name), nil
	} else {
		return NormalizeRepos(name)
	}
}

// Detect HOME path.
// If HOME environment variable is not set,
// use USERPROFILE environment variable instead.
func HomeDir() string {
	home := os.Getenv("HOME")
	if home != "" {
		return home
	}

	home = os.Getenv("USERPROFILE") // windows
	if home != "" {
		return home
	}

	panic("Couldn't look up HOME")
}

// $HOME/volt
func VoltPath() string {
	path := os.Getenv("VOLTPATH")
	if path != "" {
		return path
	}
	return filepath.Join(HomeDir(), "volt")
}

func FullReposPath(reposPath ReposPath) string {
	reposList := strings.Split(filepath.ToSlash(reposPath.String()), "/")
	paths := make([]string, 0, len(reposList)+2)
	paths = append(paths, VoltPath())
	paths = append(paths, "repos")
	paths = append(paths, reposList...)
	return filepath.Join(paths...)
}

// https://{reposPath}
func CloneURL(reposPath ReposPath) string {
	return "https://" + filepath.ToSlash(reposPath.String())
}

func Plugconf(reposPath ReposPath) string {
	filenameList := strings.Split(filepath.ToSlash(reposPath.String()+".vim"), "/")
	paths := make([]string, 0, len(filenameList)+2)
	paths = append(paths, VoltPath())
	paths = append(paths, "plugconf")
	paths = append(paths, filenameList...)
	return filepath.Join(paths...)
}

const ProfileVimrc = "vimrc.vim"
const ProfileGvimrc = "gvimrc.vim"
const Vimrc = "vimrc"
const Gvimrc = "gvimrc"

// $HOME/volt/rc/{profileName}
func RCDir(profileName string) string {
	return filepath.Join([]string{VoltPath(), "rc", profileName}...)
}

var packer = strings.NewReplacer("_", "__", "/", "_")
var unpacker1 = strings.NewReplacer("_", "/")
var unpacker2 = strings.NewReplacer("//", "_")

// Encode repos path to directory name.
// The directory name is: ~/.vim/pack/volt/opt/{name}
func EncodeReposPath(reposPath ReposPath) string {
	path := packer.Replace(reposPath.String())
	return filepath.Join(VimVoltOptDir(), path)
}

// Decode name to repos path.
// name is directory name: ~/.vim/pack/volt/opt/{name}
func DecodeReposPath(name string) ReposPath {
	name = filepath.Base(name)
	return ReposPath(unpacker2.Replace(unpacker1.Replace(name)))
}

// $HOME/volt/lock.json
func LockJSON() string {
	return filepath.Join(VoltPath(), "lock.json")
}

// $HOME/volt/config.toml
func ConfigTOML() string {
	return filepath.Join(VoltPath(), "config.toml")
}

// $HOME/volt/trx.lock
func TrxLock() string {
	return filepath.Join(VoltPath(), "trx.lock")
}

// $HOME/tmp
func TempDir() string {
	return filepath.Join(VoltPath(), "tmp")
}

// Detect vim executable path.
// If VOLT_VIM environment variable is set, use it.
// Otherwise look up "vim" binary from PATH.
func VimExecutable() (string, error) {
	var vim string
	if vim = os.Getenv("VOLT_VIM"); vim != "" {
		return vim, nil
	}
	exeName := "vim"
	if runtime.GOOS == "windows" {
		exeName = "vim.exe"
	}
	return exec.LookPath(exeName)
}

// Windows: $HOME/vimfiles
// Otherwise: $HOME/.vim
func VimDir() string {
	if runtime.GOOS == "windows" {
		return filepath.Join(HomeDir(), "vimfiles")
	} else {
		return filepath.Join(HomeDir(), ".vim")
	}
}

// (vim dir)/pack/volt
func VimVoltDir() string {
	return filepath.Join(VimDir(), "pack", "volt")
}

// (vim dir)/pack/volt/opt
func VimVoltOptDir() string {
	return filepath.Join(VimDir(), "pack", "volt", "opt")
}

// (vim dir)/pack/volt/start
func VimVoltStartDir() string {
	return filepath.Join(VimDir(), "pack", "volt", "start")
}

// (vim dir)/pack/volt/build-info.json
func BuildInfoJSON() string {
	return filepath.Join(VimVoltDir(), "build-info.json")
}

// (vim dir)/pack/volt/start/system/plugin/bundled_plugconf.vim
func BundledPlugConf() string {
	return filepath.Join(VimVoltStartDir(), "system", "plugin", "bundled_plugconf.vim")
}

// Look up vimrc path from the following candidates:
//   Windows  : $HOME/_vimrc
//              (vim dir)/vimrc
//   Otherwise: $HOME/.vimrc
//              (vim dir)/vimrc
func LookUpVimrc() []string {
	var vimrcPaths []string
	if runtime.GOOS == "windows" {
		vimrcPaths = []string{
			filepath.Join(HomeDir(), "_vimrc"),
			filepath.Join(VimDir(), "vimrc"),
		}
	} else {
		vimrcPaths = []string{
			filepath.Join(HomeDir(), ".vimrc"),
			filepath.Join(VimDir(), "vimrc"),
		}
	}
	for i := 0; i < len(vimrcPaths); {
		if !Exists(vimrcPaths[i]) {
			vimrcPaths = append(vimrcPaths[:i], vimrcPaths[i+1:]...)
			continue
		}
		i++
	}
	return vimrcPaths
}

// Look up gvimrc path from the following candidates:
//   Windows  : $HOME/_gvimrc
//              (vim dir)/gvimrc
//   Otherwise: $HOME/.gvimrc
//              (vim dir)/gvimrc
func LookUpGvimrc() []string {
	var gvimrcPaths []string
	if runtime.GOOS == "windows" {
		gvimrcPaths = []string{
			filepath.Join(HomeDir(), "_gvimrc"),
			filepath.Join(VimDir(), "gvimrc"),
		}
	} else {
		gvimrcPaths = []string{
			filepath.Join(HomeDir(), ".gvimrc"),
			filepath.Join(VimDir(), "gvimrc"),
		}
	}
	for i := 0; i < len(gvimrcPaths); {
		if !Exists(gvimrcPaths[i]) {
			gvimrcPaths = append(gvimrcPaths[:i], gvimrcPaths[i+1:]...)
			continue
		}
		i++
	}
	return gvimrcPaths
}

// Returns true if path exists, otherwise returns false.
// Existence is checked by os.Lstat().
func Exists(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}
