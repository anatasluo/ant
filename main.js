"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var electron_1 = require("electron");
var path = require("path");
var url = require("url");
var fs = require("fs");
var os = require("os");
var child_process_1 = require("child_process");
require('update-electron-app')();
var win, serve;
var args = process.argv.slice(1);
serve = args.some(function (val) { return val === '--serve'; });
var aimVersion = electron_1.app.getVersion();
var gotTheLock = electron_1.app.requestSingleInstanceLock();
var tray = null;
var torrentEngine = null;
var processExit = false;
if (gotTheLock) {
    try {
        // run torrent engine
        if (!serve) {
            runEngine();
        }
        // This method will be called when Electron has finished
        // initialization and is ready to create browser windows.
        // Some APIs can only be used after this event occurs.
        electron_1.app.on('ready', function () {
            createTray();
            createWindow();
        });
        // Quit when all windows are closed.
        electron_1.app.on('window-all-closed', function () {
            // On OS X it is common for applications and their menu bar
            // to stay active until the user quits explicitly with Cmd + Q
            if (process.platform !== 'darwin') {
                exitApp();
            }
        });
        electron_1.app.on('activate', function () {
            // On OS X it's common to re-create a window in the app when the
            // dock icon is clicked and there are no other windows open.
            if (win === null) {
                createWindow();
            }
        });
        electron_1.app.on('second-instance', function (evt, commandLine, workingDirectory) {
            console.log('Second instance');
            if (win !== null) {
                win.show();
                win.focus();
            }
        });
        // handle system restart or shutdown
        process.on('exit', function () {
            console.log('process exit');
            exitApp();
        });
    }
    catch (e) {
        console.log(e);
        // Catch Error
        // throw e;
    }
}
else {
    console.log('Only one instance can run');
    electron_1.app.quit();
}
function exitApp() {
    if (processExit === false) {
        processExit = true;
    }
    else {
        return;
    }
    console.log('Close everything left');
    if (win !== null) {
        win.destroy();
    }
    if (torrentEngine !== null) {
        torrentEngine.kill();
    }
    if (tray !== null) {
        tray.destroy();
    }
    if (electron_1.app !== null) {
        electron_1.app.quit();
    }
}
function runEngine() {
    console.log('Engine running');
    var userData = electron_1.app.getPath('userData');
    var systemVersion = os.platform();
    var cmdPath;
    if (systemVersion === 'win32') {
        cmdPath = '/ant_' + aimVersion + '.exe';
    }
    else {
        cmdPath = '/ant_' + aimVersion;
    }
    // Copy file from app.asar to user data
    // Do nothing if find needed binary file
    if (!fs.existsSync(userData + cmdPath)) {
        console.log('version update');
        copyFile(electron_1.app.getAppPath() + '/torrent' + cmdPath, userData + cmdPath);
        copyFile(electron_1.app.getAppPath() + '/torrent' + '/tracker.txt', userData + '/tracker.txt');
        copyFile(electron_1.app.getAppPath() + '/torrent' + '/config.toml', userData + '/config.toml');
    }
    fs.chmodSync(userData + cmdPath, '0555');
    torrentEngine = child_process_1.execFile(userData + cmdPath, { cwd: userData }, function (error, stdout, stderr) {
        if (error) {
            console.error("Exec Failed: " + error);
            return;
        }
        console.log("stdout: " + stdout);
        console.log("stderr: " + stderr);
    });
    torrentEngine.stdout.on('data', function (data) {
        console.log('stdout: ' + data);
    });
    torrentEngine.stderr.on('data', function (data) {
        console.log('stderr: ' + data);
    });
    torrentEngine.on('close', function (code) {
        console.log('out code：' + code);
        exitApp();
    });
}
function createTray() {
    var trayIcon = path.join(electron_1.app.getAppPath(), 'logo.png');
    var nimage = electron_1.nativeImage.createFromPath(trayIcon);
    tray = new electron_1.Tray(nimage);
    tray.setToolTip('ANT Downloader');
    var contextMenu = electron_1.Menu.buildFromTemplate([
        {
            label: 'Show', click: function () {
                win.show();
            }
        },
        {
            label: '',
            type: 'separator'
        },
        {
            label: 'Quit', click: function () {
                exitApp();
            }
        },
    ]);
    tray.setTitle('ANT Downloader');
    tray.setContextMenu(contextMenu);
    // Not work for linux
    tray.on('click', function () {
        win.isVisible() ? win.hide() : win.show();
    });
}
function createWindow() {
    var electronScreen = electron_1.screen;
    var size = electronScreen.getPrimaryDisplay().workAreaSize;
    // Create the browser window.
    // src is not a valid directory after package
    win = new electron_1.BrowserWindow({
        width: size.width * 0.8,
        height: size.height * 0.75,
        minWidth: size.width * 0.65,
        minHeight: size.height * 0.7,
        title: 'ANT Downloader',
        icon: path.join(electron_1.app.getAppPath(), 'logo.png'),
        autoHideMenuBar: true,
        titleBarStyle: 'hidden',
    });
    if (serve) {
        require('electron-reload')(__dirname, {
            electron: require(__dirname + "/node_modules/electron")
        });
        win.loadURL('http://localhost:4200');
    }
    else {
        win.loadURL(url.format({
            pathname: path.join(__dirname, 'dist/index.html'),
            protocol: 'file:',
            slashes: true
        }));
    }
    if (serve) {
        win.webContents.openDevTools();
    }
    win.on('minimize', function (evt) { });
    win.on('close', function (evt) {
        win.hide();
        win.setSkipTaskbar(true);
        evt.preventDefault();
    });
    win.on('closed', function (evt) {
        console.log('App quit now');
        exitApp();
    });
    win.webContents.session.on('will-download', function (event, item, webContents) {
        var filePath = electron_1.app.getPath('downloads') + '/' + item.getFilename();
        console.log(filePath);
        item.setSavePath(filePath);
        item.once('done', function (evt, state) {
            if (state === 'completed') {
                console.log('Download successfully');
                win.webContents.send('torrentDownload', filePath);
            }
            else {
                console.log("Download failed: " + state);
            }
        });
    });
}
function copyFile(src, dist) {
    if (!fs.existsSync(dist)) {
        fs.writeFileSync(dist, fs.readFileSync(src));
    }
}
//# sourceMappingURL=main.js.map