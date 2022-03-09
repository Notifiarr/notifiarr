This folder is special and can contain extra files for your application packages.
If you want to add a package to a deb package, create a folder named `deb` and put
your files into it. Same for rpm.

For instance, if you want to add a file in ``/etc/sudoers.d`, do something like this:

```
mkdir init/deb/etc/sudoers.d
echo "#sudoers content" > init/deb/etc/sudoers.d/my-sudoers
```

Same thing for rpm packages:

```
mkdir init/rpm/etc/sudoers.d
echo "#sudoers content" > init/rpm/etc/sudoers.d/my-sudoers
