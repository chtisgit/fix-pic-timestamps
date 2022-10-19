# fix-pic-timestamps

Go through files and look for timestamps in their names, then apply those timestamps
to the file's last modified timestamp. This is essential for sorting files by
timestamp. You can sort these regularly named files by name, you say? That only works
if all files in the directory follow the same naming scheme.

Example filenames:

```
IMG_20220804_200011.jpg
P_20170216_115449.jpg
VID_20220928_084904.mp4
```

The timestamp pattern is hard-coded in the program at the moment.

I used **Go** to do this, because it can be easily cross-compiled for phones.

## Program Usage

When you're new to this program, always run with `-n` (dry-run) flag first.
It will not act, but only show the changes it would perform.

```
Usage of ./fix-pic-timestamps:
  -i    interactive mode
  -ignore-errors
        keep on running when errors occur
  -n    dry run (enables verbose)
  -off int
        file timestamps can be off by this many ms without being touched by this tool (default 2000)
  -r    recursively process directories
  -tz string
        set timezone to use (if not set, uses system timezone)
  -v    verbose
```
