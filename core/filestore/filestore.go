package filestore

import (
	"onlinetools/core/control"
	"os/exec"
	"path"
	"strings"
)

type FileStore struct {
	RemoteName string
	UrlHeader  string
	cmddir     string
}

/*
this func can access remote store.
aliyun is mounted at liyun
other
*/
func NewAlistAccesser() *FileStore {
	return &FileStore{
		RemoteName: "alist:",
		UrlHeader:  "http://8.142.131.147:5244/d/",
		cmddir:     path.Join(control.PreToolsPath, "rclone", "rclone"),
	}
}

func exeCMD(execstr string, args ...string) (string, error) {
	cmd := exec.Command(execstr, args...)
	outStr, err := cmd.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(outStr), nil
}

/*
Copy the source to the destination.  Does not transfer files that are
identical on source and destination, testing by size and modification
time or MD5SUM.  Doesn't delete files from the destination. If you
want to also delete files from destination, to make it match source,
use the [sync](/commands/rclone_sync/) command instead.

Note that it is always the contents of the directory that is synced,
not the directory itself. So when source:path is a directory, it's the
contents of source:path that are copied, not the directory name and
contents.

To copy single files, use the [copyto](/commands/rclone_copyto/)
command instead.

If dest:path doesn't exist, it is created and the source:path contents
go there.
*/
func (f *FileStore) Copy(sourcepath string, destpath string) (string, error) {
	return exeCMD(f.cmddir, "copy", sourcepath, f.RemoteName+destpath)
}

func (f *FileStore) CopyTo(sourcepath string, destpath string) (string, error) {
	return exeCMD(f.cmddir, "copyto", sourcepath, f.RemoteName+destpath)
}

/*
Sync the source to the destination, changing the destination
only.  Doesn't transfer files that are identical on source and
destination, testing by size and modification time or MD5SUM.
Destination is updated to match source, including deleting files
if necessary (except duplicate objects, see below). If you don't
want to delete files from destination, use the
[copy](/commands/rclone_copy/) command instead.

If dest:path doesn't exist, it is created and the source:path contents
go there.
*/
func (f *FileStore) Sync(sourcepath string, destpath string) (string, error) {
	return exeCMD(f.cmddir, "sync", sourcepath, f.RemoteName+destpath)
}

/*
Remove the files in path.  Unlike [purge](/commands/rclone_purge/) it
obeys include/exclude filters so can be used to selectively delete files.

`rclone delete` only deletes files but leaves the directory structure
alone. If you want to delete a directory and all of its contents use
the [purge](/commands/rclone_purge/) command.

If you supply the `--rmdirs` flag, it will remove all empty directories along with it.
You can also use the separate command [rmdir](/commands/rclone_rmdir/) or
[rmdirs](/commands/rclone_rmdirs/) to delete empty directories only.

For example, to delete all files bigger than 100 MiB, you may first want to
check what would be deleted (use either):

	rclone --min-size 100M lsl remote:path
	rclone --dry-run --min-size 100M delete remote:path

Then proceed with the actual delete:

	rclone --min-size 100M delete remote:path

That reads "delete everything with a minimum size of 100 MiB", hence
delete all files bigger than 100 MiB.

Usage:

	delete remote:path [flags]
*/
func (f *FileStore) Delete(destpath string) (string, error) {
	return exeCMD(f.cmddir, "delete", "--rmdirs", f.RemoteName+destpath)
}

func (f *FileStore) Lsd(path string) ([]string, error) {
	files, err := exeCMD(f.cmddir, "lsd", f.RemoteName+path)
	if err == nil {
		filenames := strings.Split(files, "\n")
		for i := range filenames {
			filenames[i] = f.UrlHeader + filenames[i]
		}
		return filenames, err
	}
	return nil, err

}

func (f *FileStore) GetUrl(destpath string) ([]string, error) {
	files, err := exeCMD(f.cmddir, "lsf", f.RemoteName+destpath)
	if err == nil {
		filenames := strings.Split(files, "\n")
		for i := range filenames {
			filenames[i] = f.UrlHeader + strings.TrimSpace(filenames[i])
		}
		return filenames, err
	}
	return nil, err
}
