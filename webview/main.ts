import { app, BrowserWindow, screen, Menu, Tray } from 'electron';
import * as path from 'path';
import * as url from 'url';

let win, serve;
const args = process.argv.slice(1);
serve = args.some(val => val === '--serve');

// let tray = null;

function createWindow() {

  const electronScreen = screen;
  const size = electronScreen.getPrimaryDisplay().workAreaSize;

  // Create the browser window.
  win = new BrowserWindow({
    width: size.width * 0.65,
    height: size.height * 0.7,
    minWidth: size.width * 0.6,
    minHeight: size.height * 0.6,
    title: 'ANT Downloader',
    icon: path.join(__dirname, 'src/assets/tray.png'),
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

  if (serve || true) {
    win.webContents.openDevTools();
  }

  // Emitted when the window is closed.
  win.on('closed', () => {
    // Dereference the window object, usually you would store window
    // in an array if your app supports multi windows, this is the time
    // when you should delete the corresponding element.
    win = null;
    // tray.destroy();
  });

  // tray = new Tray(path.join(__dirname, 'src/assets/tray.png'));
  // const contextMenu = Menu.buildFromTemplate([
  //   { label: '新建下载', type: 'normal' },
  //   { label: '', type: 'separator' },
  //   { label: '全部开始', type: 'normal' },
  //   { label: '全部暂停', type: 'normal' },
  // ]);
  // tray.setToolTip('ANT Downloader');
  // tray.setContextMenu(contextMenu);

}

try {

  // This method will be called when Electron has finished
  // initialization and is ready to create browser windows.
  // Some APIs can only be used after this event occurs.
  app.on('ready', createWindow);

  // Quit when all windows are closed.
  app.on('window-all-closed', () => {
    // On OS X it is common for applications and their menu bar
    // to stay active until the user quits explicitly with Cmd + Q
    if (process.platform !== 'darwin') {
      app.quit();
    }
  });

  app.on('activate', () => {
    // On OS X it's common to re-create a window in the app when the
    // dock icon is clicked and there are no other windows open.
    if (win === null) {
      createWindow();
    }
  });

} catch (e) {
  // Catch Error
  // throw e;
}
