import { app, BrowserWindow, screen, Menu, Tray, ipcMain, nativeImage } from 'electron';
import * as path from 'path';
import * as url from 'url';
import * as fs from 'fs';
import * as os from 'os';
import { execFile, ChildProcess } from 'child_process';

require('update-electron-app')();

let win: BrowserWindow, serve;
const args = process.argv.slice(1);
serve = args.some(val => val === '--serve');

const aimVersion = app.getVersion();
const gotTheLock = app.requestSingleInstanceLock();

let tray: Tray = null;
let torrentEngine: ChildProcess = null;
let processExit: Boolean = false;

const userData = app.getPath('userData');

if (gotTheLock) {
    try {
        // run torrent engine
        if (!serve) {
            runEngine();
        }
        // This method will be called when Electron has finished
        // initialization and is ready to create browser windows.
        // Some APIs can only be used after this event occurs.
        app.on('ready', () => {
            createTray();
            createWindow();
        });
        // Quit when all windows are closed.
        app.on('window-all-closed', () => {
            // On OS X it is common for applications and their menu bar
            // to stay active until the user quits explicitly with Cmd + Q
            if (process.platform !== 'darwin') {
                exitApp();
            }
        });

        app.on('activate', () => {
            // On OS X it's common to re-create a window in the app when the
            // dock icon is clicked and there are no other windows open.
            if (win === null) {
                createWindow();
            }
        });

        app.on('second-instance',  (evt, commandLine, workingDirectory) => {
            console.log('Second instance');
            if (win !== null) {
                win.show();
                win.focus();
            }
        });

        // handle system restart or shutdown
        process.on('exit', function() {
            console.log('process exit');
            exitApp();
        });

    } catch (e) {
        console.log(e);
        // Catch Error
        // throw e;
    }
} else {
    console.log('Only one instance can run');
    app.quit();
}

function exitApp() {
    if (processExit === false) {
        processExit = true;
    } else {
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
    if (app !== null) {
        app.quit();
    }
}
function runEngine() {
    console.log('Engine running');
    const systemVersion = os.platform();
    let cmdPath;
    if (systemVersion === 'win32') {
        cmdPath = '/ant_' + aimVersion + '.exe';
    } else {
        cmdPath = '/ant_' + aimVersion;
    }
    // Copy file from app.asar to user data
    // Do nothing if find needed binary file
    if (!fs.existsSync(userData + cmdPath)) {
        console.log('version update');
        copyFile(app.getAppPath() + '/torrent' + cmdPath, userData + cmdPath);
        copyFile(app.getAppPath() + '/torrent' + '/tracker.txt', userData + '/tracker.txt');
    }

    // restore broken setting file
    restoreSettingFile(app.getAppPath() + '/torrent' + '/config.toml', userData + '/config.toml');
    fs.chmodSync(userData + cmdPath, '0555');

    torrentEngine = execFile(userData + cmdPath, {cwd: userData}, (error, stdout, stderr) => {
        if (error) {
            console.error(`Exec Failed: ${error}`);
            return;
        }
        console.log(`stdout: ${stdout}`);
        console.log(`stderr: ${stderr}`);
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
    const trayIcon = path.join(app.getAppPath(), 'logo.png');
    const nimage = nativeImage.createFromPath(trayIcon);
    tray = new Tray(nimage);
    tray.setToolTip('ANT Downloader');
    const contextMenu = Menu.buildFromTemplate([
        {
            label: 'Show', click: () => {
                win.show();
            }
        },
        {
            label: '',
            type: 'separator'
        },
        {
            label: 'Quit', click: () => {
                exitApp();
            }
        },
    ]);
    tray.setTitle('ANT Downloader');
    tray.setContextMenu(contextMenu);

    // Not work for linux
    tray.on('click', () => {
        win.isVisible() ? win.hide() : win.show();
    });
}

function createWindow() {

    const electronScreen = screen;
    const size = electronScreen.getPrimaryDisplay().workAreaSize;

    // Create the browser window.
    // src is not a valid directory after package
    win = new BrowserWindow({
        width: size.width * 0.8,
        height: size.height * 0.75,
        minWidth: size.width * 0.65,
        minHeight: size.height * 0.7,
        title: 'ANT Downloader',
        icon: path.join(app.getAppPath(), 'logo.png'),
        autoHideMenuBar: true,
        titleBarStyle: 'hidden',
    });

    if (serve) {
        require('electron-reload')(__dirname, {
            electron: require(`${__dirname}/node_modules/electron`)
        });
        win.loadURL('http://localhost:4200');
    } else {
        win.loadURL(url.format({
            pathname: path.join(__dirname, 'dist/index.html'),
            protocol: 'file:',
            slashes: true
        }));
    }

    if (serve) {
        win.webContents.openDevTools();
    }

    win.on('minimize', (evt) => {});

    win.on('close', (evt) => {
        win.hide();
        win.setSkipTaskbar(true);
        evt.preventDefault();
    });

    win.on('closed', (evt) => {
        console.log('App quit now');
        exitApp();
    });

    win.webContents.session.on('will-download', (event, item, webContents) => {
        const filePath = app.getPath('downloads') + '/' + item.getFilename();
        console.log(filePath);
        item.setSavePath(filePath);
        item.once('done', (evt, state) => {
            if (state === 'completed') {
                console.log('Download successfully');
                win.webContents.send('torrentDownload', filePath);
            } else {
                console.log(`Download failed: ${state}`);
            }
        });
    });

}
function restoreSettingFile(src, dst) {
    const srcLines = fs.readFileSync(src).toString().split('\n').length;
    const dstLines = fs.readFileSync(dst).toString().split('\n').length;
    if (srcLines !== dstLines) {
        console.log('restore broken setting file.');
        fs.writeFileSync(dst, fs.readFileSync(src));
    }
}

function copyFile(src, dst) {
    if (!fs.existsSync(dst)) {
        fs.writeFileSync(dst, fs.readFileSync(src));
    }
}
