
![](./src/assets/READEME/logoAndID.png)

## ANT Downloader

[![Build Status](https://travis-ci.com/anatasluo/ant.svg?branch=master)](https://travis-ci.com/anatasluo/ant)
[![Stable Version](https://img.shields.io/badge/version-1.1.0-blueviolet.svg)](https://img.shields.io/badge/version-1.1.0-blueviolet.svg)
[![License: MPL 2.0](https://img.shields.io/badge/License-MPL%202.0-brightgreen.svg)](https://opensource.org/licenses/MPL-2.0)

### [English](README.md) | [中文](README_zh.md)

> ANT Downloader is a BitTorrent Client developed by golang, angular 7, and electron. ANT focuses on supplying best user experience for torrent resource download with least system resource.  

If you like this application, please consider give a star for this project.

## Why you should consider ANT Downloader:
- a BitTorrent client for all platforms
- a BitTorrent client with beautiful UI
- a BitTorrent client with low resource occupancy.
- a BitTorrent client with rich set of functions like steaming video while downloading
- a BitTorrent client which only concentrates on resource download
- ANT uses many public torrent services to supply better user experience like
  - [itorrents](https://itorrents.org/)
  - [trackerslist](https://github.com/ngosang/trackerslist)
  - [thepiratebay](https://www.thepiratebay.org/)
  - ...

## Considering features in following version:
- [ ] Download and steam selected file (Current version will download all files in one torrent and only steam the biggest file.)
- [ ] Support different UI themes
- [ ] Support more download methods like baiduyun, webTorrent
- [ ] Automatically open local player to play video.
- [ ] Control ANT Downloader from remote machine.

## Architecture:
![](./src/assets/READEME/architecture.jpg)

## Preview:
+ ### Add torrent download task
![](./src/assets/READEME/task.gif)
--------------

+ ### Steaming video while downloading
![](./src/assets/READEME/steaming.png)
--------------


## Get Started

You can download packaged binary file directly from [Release](https://github.com/anatasluo/ant/releases)

You can also build project with one of following cmd, and it depends on your system:
```
npm run electron:linux
```

```
npm run electron:windows
```

```
npm run electron:mac
```

More npm usage is described in package.json, make sure your system has following dependences
+ node >= 11.0.x
+ golang >= 1.10.x

# Contact me
You can send emails to luolongjuna@gmail.com.
