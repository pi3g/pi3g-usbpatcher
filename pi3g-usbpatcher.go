package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

const version = "0.0.3"

const mountPath = "/mnt"
const logPath = "/var/log/pi3g-usbpatcher"
const updaterPath = "/usr/bin/pi3g-usbpatcher"
var filenameRegexp = regexp.MustCompile(`^pi3g-patch-.+?.(tgz|tar.gz)$`)

const mountBin = "/bin/mount"
const umountBin = "/bin/umount"
const tarBin = "/bin/tar"
const haltBin = "/sbin/halt"

// mount mounts the device at devname to mountPath
func mount(devname string) error {
	return exec.Command(mountBin, devname, mountPath).Run()
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
	// start logging
	logfile, err := os.OpenFile(logPath, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Println(err)
		log.Println("Falling back to stdout logging")
	} else {
		log.SetOutput(logfile)
	}
	defer logfile.Close()
	log.Println("Device plugged in, running updater version ", version)

	devname := os.Getenv("DEVNAME")
	if devname == "" {
		log.Fatalln("DEVNAME not set")
	}
	log.Println("Device found: ", devname)

	err = mount(devname)
	if err != nil {
		log.Fatalln("mount: ", err)
	}

	patchFile := findPatchFile()
	if patchFile == "" {
		err = umount(devname)
		if err != nil {
			log.Println("umount: ", err)
		}
		log.Fatalln("No patch on drive")
	}
	patchFile = mountPath + "/" + patchFile
	log.Println("Patch file found: ", patchFile)

	out, err := archiveList(patchFile)
	if err != nil {
		err = umount(devname)
		if err != nil {
			log.Println("umount: ", err)
		}
		log.Fatalln("tar list: ", err)
	}
	log.Printf("The following files will be updated:\n%s", out)

	// unlink this binary if it is going to be replaced
	re := regexp.MustCompile(`^[^/]*` + updaterPath + `$`)
	for _, line := range strings.Fields(out) {
		if re.MatchString(line) {
			log.Println("Warning: Contains update for updater!")
			err := os.Remove(updaterPath)
			if err != nil {
				err = umount(devname)
				if err != nil {
					log.Println("umount: ", err)
				}
				log.Println("This is really bad!")
				log.Fatalln("remove: ", err)
			}
			break
		}
	}

	err = archiveExtract(patchFile)
	if err != nil {
		err = umount(devname)
		if err != nil {
			log.Println("umount: ", err)
		}
		log.Fatalln("tar extract: ", err)
	}

	// we couldn't defer this because main usually doesn't return
	err = umount(devname)
	if err != nil {
		log.Println("umount: ", err)
	}

	log.Println("Shutting down")
	err = halt()
	if err != nil {
		log.Fatalln("halt: ", err)
	}
}
