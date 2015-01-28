# Patchfiles

The Program scans the flash drive for files of the form `pi3g-patch-*.tgz`. If
an archive is found, its content is extracted to `/` and existing files are
replaced with their newer versions.

Patchfiles are gzip compressed tar archives and can be created with standard
tools. This example will create a file `helloworld.txt` in the pi users home
directory:

    mkdir -p patchdir/home/pi/
    touch patchdir/home/pi/helloworld.txt
    tar czvf pi3g-patch-test.tar.gz -C patchdir .

# Building

To build for the Raspberry Pi run:

    GOARCH=arm GOARM=5 go build github.com/pi3g/pi3g-netconf

For logging of debug messages add `-tags debug` like this:

    GOARCH=arm GOARM=5 go build -tags debug github.com/pi3g/pi3g-netconf

# TODO

Planned features:

- Configuration file
- Package signing
- Deleting files
- Installing packages
