To make a new default template, run this:

```
# Build the binary.
make
# Re-write Template
./notifiarr -c examples/notifiarr.conf.example --write example
# Check it out
git diff examples
```
