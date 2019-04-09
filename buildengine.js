// This script is used to generate binary file from backend directory
// Out directory is set to torrent, this should not change
const os = require('os');
const argv = require('optimist')
    .default('system', os.platform())
    .default('arch', os.arch())
    .argv;
const _ = require('lodash');
const util = require('util');
const exec = require('child_process').exec;
const fs = require('fs');
const path = require('path');

const aimVersion = process.env.npm_package_version;
let aimArch, aimPlatform, buildCmd;
let sourcePath, destPath;

// handle arch argv
if (_.startsWith(argv.arch, 'arm'))
{
    aimArch = 'arm';
}else if (_.startsWith(argv.arch, 'x32') || _.startsWith(argv.arch, '386'))
{
    aimArch = '386';
}else if (_.startsWith(argv.arch, 'x64') || _.startsWith(argv.arch, 'amd64'))
{
    aimArch = 'amd64';
}else {
    console.log("Not support such platform");
    process.exit(1);
}

// hadnle system argv
if (argv.system === 'darwin' || argv.system === 'linux')
{
    aimPlatform = argv.system;
    sourcePath = './backend/ant';
    destPath = './torrent/ant_' + aimVersion;
}else if (argv.system === 'win32' || argv.system === 'windows')
{
    aimPlatform = 'windows';
    sourcePath = './backend/ant.exe';
    destPath = './torrent/ant_' + aimVersion + '.exe';
}else {
    console.log("Not support such platform");
    process.exit(1);
}

let currentOS = os.platform();

if (currentOS === 'linux' || currentOS === 'darwin')
{
    buildCmd = util.format('CGO_ENABLED=0 GOOS=%s GOARCH=%s go build ant.go', aimPlatform, aimArch);
}else if (currentOS === 'win32')
{
    buildCmd = util.format('SET CGO_ENABLED=0 SET GOOS=%s SET GOARCH=%s go build ant.go', aimPlatform, aimArch);
}else
{
    console.log("Not support such platform");
    process.exit(1);
}

console.log(buildCmd);

// clear out directory
removeDir('./torrent');
fs.mkdirSync('./torrent');
// build engine
exec(buildCmd, {
    "cwd": "./backend"
}, function(error, stdout, stderr){
    if(error) {
        console.log("Find error in building engine");
        console.log(error);
        console.log(stderr);
        process.exit(1);
    }
    console.log(stdout);
    fs.renameSync(sourcePath, destPath);
    copyFile('./backend/config.toml', './torrent/config.toml');
    copyFile('./backend/tracker.txt', './torrent/tracker.txt');
});

function copyFile(src, dist) {
    if (!fs.existsSync(dist)) {
        fs.writeFileSync(dist, fs.readFileSync(src));
    }
}

function removeDir(dir) {
    let files = fs.readdirSync(dir);
    for(let i=0;i<files.length;i++){
        let newPath = path.join(dir,files[i]);
        let stat = fs.statSync(newPath);
        if(stat.isDirectory()){
            removeDir(newPath);
        }else {
            fs.unlinkSync(newPath);
        }
    }
    fs.rmdirSync(dir);
}




