# utah

[RIP Utah Phillips](http://thelongmemory.com/)

A slow cooking attempt to build a less aggravating alternative to [Vagrant](http://vagrantup.com/),
and maybe a more respectful one.

# Building

I recommend [gb](http://getgb.io/). Install it, and run

    gb build all

You can also build it with the standard tools:

    export GOPATH=`pwd`
    cd src/cmd/utah
    go build

# Third Party Binaries

Depends on Virtualbox and qemu-img being installed and available.
