# Sandbox

A lightweight tool for running commands in isolated Linux namespaces.

Sandbox uses Linux container primitives to run processes in isolated
environments with a builtin minimal filesystem.

> [!IMPORTANT]
> This tool uses **Linux namespaces** so the tool can only run on Linux systems ;)

## Why use this if I already have Docker?

Docker does much more than spawn a process inside a container. It manages
networking, images, restarts, resources, and more. Sandbox is intended to
be a simple mechanism for running a process in an isolated environment.

## The process doesn't have network access

That is not a bug, it is a feature. Sandbox only creates the namespaces in
which the process runs. If you want networking, you must configure it. If
you want configuration files, you must provide them.

## What this is not

Sandbox **is not** a replacement for Docker, containerd, or any other
container runtime. It is a tool for experimenting with processes in
isolated environments and learning how Linux namespaces work.

Use disposable root filesystems for destructive experiments.

## Attributions

The idea for this tool came to me while I was choking on my own saliva.
I spent countless hours trying to understand **namespaces**, and this tool
is a compilation of everything I learned. Nonetheless, I took inspiration
from these two YouTube videos.

Huge thanks to Liz Rice and Gerlof Langeveld for sharing the knowledge :)

<p align="center">
   <a href="https://www.youtube.com/watch?v=BM3aH-wultc">
      <img width="240" height="180" alt="YouTube video" title="Containers – A Look Under the Hood - Gerlof Langeveld, AT Computing" src="https://img.youtube.com/vi/BM3aH-wultc/hqdefault.jpg"/>
   </a>
   <a href="https://www.youtube.com/watch?v=jeTKgAEyhsA">
      <img width="240" height="180" alt="YouTube video" title="Rootless Containers from Scratch - Liz Rice, Aqua Security" src="https://img.youtube.com/vi/jeTKgAEyhsA/hqdefault.jpg"/>
   </a>
</p>

## How does it work?

In Linux, "containers" isn't a thing — there's no feature you enable or
disable called "containers". **If** you watched the videos in the
attributions (which I highly suggest doing) you may now know that:

> A container is a process that **_unshares_** the selected inherited
> namespaces from the parent process.

And a **namespace** is an isolation mechanism that allows processes to have
a restricted _view_ or _access_ to a set of Linux features. What kind of
features? i.e. process id table, users mappings, network stack, etc.

So, at a very high level, what this tool does is run a process _unsharing_
the selected inherited namespaces from the parent.

## Examples

### Unshare network namespace

If you run the command `ip addr` on your system, you may have multiple IP
addresses and multiple interfaces. Check your current network namespace by
running this command:

```sh
readlink /proc/$$/ns/net
net:[4026531833]
```

If you run a shell in your current shell, that _child_ process will inherit
the same namespace. Try yourself:

```sh
bash
readlink /proc/$$/ns/net
net:[4026531833]
```

Now, what we can do, is **unshare** the network namespace for the new shell

```sh
sudo unshare --net bash
readlink /proc/$$/ns/net
net:[4026533612]
```

And as simple as that we have isolated a process from our network namespace.

```sh
ip addr
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
```

As you can see, in this new namespace we do not have any network interface.
Now you know how sandbox works :).

```sh
readlink /proc/$$/ns/net
net:[4026531833]

sb run -A sh

/ # readlink /proc/$$/ns/net
net:[4026533626]

/ # ip addr
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
```

Now that we have seen briefly how `unshare` works, what makes this tool
something useful and not an elaborated Go wrapper? Good question.

Sandbox comes with a minimal builtin filesystem from alpine that allows
you to unshare the `mount (mnt)` namespace without worrying about it.
