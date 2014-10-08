package main

import (
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var version = "1.0.2"

const mountPath = "/mnt"
const updaterPath = "/usr/bin/pi3g-usbpatcher"

var filenameRegexp = regexp.MustCompile(`^pi3g-patch-.+?.(tgz|tar.gz)$`)

const mountBin = "/bin/mount"
const umountBin = "/bin/umount"
const tarBin = "/bin/tar"
const haltBin = "/sbin/halt"

// mount mounts the device at devname to mountPath
func mount(devname string) error {
	return exec.Command(mountBin, "-r", devname, mountPath).Run()
}

// umount unmounts the device at devname
func umount(devname string) error {
	return exec.Command(umountBin, devname).Run()
}

// archiveList lists contents of archive
func archiveList(archive string) (string, error) {
	out, err := exec.Command(tarBin, "tf", archive, "--strip-components=1").Output()
	return string(out), err
}

// archiveExtract extracts archive to "/" overwriting if necessary
func archiveExtract(archive string) error {
	return exec.Command(tarBin, "xzf", archive, "-C", "/",
		"--strip-components=1", "--overwrite").Run()
}

// findPatchFile searches mountPath for a patchFile candidate
func findPatchFile() string {
	files, _ := ioutil.ReadDir(mountPath)
	for _, f := range files {
		if filenameRegexp.MatchString(f.Name()) {
			return f.Name()
		}
	}
	return ""
}

// halt shuts down the raspi
func halt() error {
	return exec.Command(haltBin).Run()
}

func main() {
	debug("Device plugged in, running updater version ", version)

	devname := os.Getenv("DEVNAME")
	if devname == "" {
		debug("DEVNAME not set")
		os.Exit(1)
	}
	debug("Device found: ", devname)

	err := mount(devname)
	if err != nil {
		debug("mount: ", err)
		os.Exit(1)
	}

	patchFile := findPatchFile()
	if patchFile == "" {
		err = umount(devname)
		if err != nil {
			debug("umount: ", err)
		}
		debug("No patch on drive")
		os.Exit(1)
	}
	patchFile = mountPath + "/" + patchFile
	debug("Patch file found: ", patchFile)

	out, err := archiveList(patchFile)
	if err != nil {
		err = umount(devname)
		if err != nil {
			debug("umount: ", err)
		}
		debug("tar list: ", err)
		os.Exit(1)
	}
	debugf("The following files will be updated:\n%s", out)

	// unlink this binary if it is going to be replaced
	re := regexp.MustCompile(`^[^/]*` + updaterPath + `$`)
	for _, line := range strings.Fields(out) {
		if re.MatchString(line) {
			debug("Warning: Contains update for updater!")
			err := os.Remove(updaterPath)
			if err != nil {
				err = umount(devname)
				if err != nil {
					debug("umount: ", err)
				}
				debug("This is really bad!")
				debug("remove: ", err)
				os.Exit(1)
			}
			break
		}
	}

	err = archiveExtract(patchFile)
	if err != nil {
		err = umount(devname)
		if err != nil {
			debug("umount: ", err)
		}
		debug("tar extract: ", err)
		os.Exit(1)
	}

	// we couldn't defer this because main usually doesn't return
	err = umount(devname)
	if err != nil {
		debug("umount: ", err)
	}

	debug("Shutting down")
	err = halt()
	if err != nil {
		debug("halt: ", err)
		os.Exit(1)
	}
}
