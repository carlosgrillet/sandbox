# Sandbox

A lightweight tool for running commands in isolated Linux namespaces.

Sandbox uses Linux container primitives to run processes in isolated
environments.

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
