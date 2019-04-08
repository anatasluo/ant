import {Component, OnInit, OnDestroy } from '@angular/core';
import {TorrentService} from '../../providers/torrent.service';
import * as _ from 'lodash';
import {Torrent} from '../../classes/torrent';
import {ActivatedRoute} from '@angular/router';
import { ipcRenderer, remote, screen, Menu, shell } from 'electron';
import * as fs from 'fs';
import * as path from 'path';

import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import { ConfigService } from '../../providers/config.service';
import { MessagesService } from '../../providers/messages.service';

import { magnetDecode } from '@ctrl/magnet-link';

let globalTorrents: Torrent[];
let ws: WebSocket;
let torrentFile: File;
let currentMagnet: string;

@Component({
  selector: 'app-local-download',
  templateUrl: './local-download.component.html',
  styleUrls: ['./local-download.component.scss']
})

export class LocalDownloadComponent implements OnInit, OnDestroy {

  private aviImg = require('../../../assets/file_type/zip.png');
  torrents: Torrent[];
  // See GetTrueFromSelected
  selectedTorrent: Torrent;
  status: string;
  webview: any;
  rightMenu: Menu;
  constructor(private torrentService: TorrentService,
              private configService: ConfigService,
              private route: ActivatedRoute,
              private modalService: NgbModal,
              private messagesService: MessagesService,
  ) { }
  ngOnInit() {
    const tmpThis = this;
    currentMagnet = undefined;
    this.status = this.route.snapshot.url[0].path;
    this.getTorrents();
    if (ws === null || ws === undefined) {
      ws = this.generateOneWS();
    }
    this.webview = document.querySelector('webview');
    // @ts-ignore
    this.webview.addEventListener('dom-ready', () => {
      tmpThis.messagesService.add('itorrents service ready');
    });

    ipcRenderer.on('torrentDownload', (event, arg) => {
      tmpThis.messagesService.add('get torrent file from itorrents successfully');
      // const tmpMagnet = currentMagnet;
      const filePath: string = arg;
      if (_.endsWith(filePath, 'torrent')) {
        currentMagnet = undefined;
        this.getFileFromURL(filePath);
      } else {
        tmpThis.messagesService.add('Failed to get torrent file from itorrents');
      }
    });

    ipcRenderer.on('torrentLoaded', () => {
      this.addOneTorrentService(torrentFile);
    });

    this.rightMenu = new remote.Menu();
    this.rightMenu.append(new remote.MenuItem({
      label: 'Start task', click() {
        tmpThis.startOneTorrent();
      }
    }));

    // this.rightMenu.append(new remote.MenuItem({
    //   type: 'separator'
    // }));
    this.rightMenu.append(new remote.MenuItem({
      label: 'Stop task', click() {
        tmpThis.stopOneTorrent();
      }
    }));

    this.rightMenu.append(new remote.MenuItem({
      label: 'Delete task', click() {
        tmpThis.delOneTorrent();
      }
    }));

    this.rightMenu.append(new remote.MenuItem({
      label: 'Play while downloading', click() {
        tmpThis.showPlay();
      }
    }));

    this.rightMenu.append(new remote.MenuItem({
      label: 'Show in directory', click() {
        tmpThis.openInDirectory();
      }
    }));
  }

  ngOnDestroy() {
    ipcRenderer.removeAllListeners('torrentDownload');
    ipcRenderer.removeAllListeners('torrentLoaded');
    // ws.close();
  }

  private getFileFromURL(filePath: string) {
    const fileData = fs.createReadStream(filePath);
    const chunks = [];
    fileData.on('data', function (chunk) {
      chunks.push(chunk);
    });
    fileData.on('end', function () {
      const b: any = new Blob(chunks);
      b.lastModifiedDate = new Date();
      b.name = path.basename(filePath);
      torrentFile = <File>b;
      remote.getCurrentWebContents().send('torrentLoaded');
    });
  }

  private sendMagnet(magnet: string) {
    magnet = _.trim(magnet);
    this.torrentService.sendMagnet(magnet)
        .subscribe((IsAdded: boolean) => {
          if (IsAdded) {
            // location.reload();
            this.getTorrents();
          }
        }, error => {
          console.log(error);
        });
  }

  private generateOneWS(): WebSocket {

    const tmpThis = this;
    const tmpWS = new WebSocket(this.configService.wsBaseUrl);
    tmpWS.onopen = function(evt: any) {
      tmpThis.messagesService.add('create websocket');
    };
    tmpWS.onclose = function(evt: any) {
      tmpThis.messagesService.add('close websocket');
      ws.close();
      setTimeout(() => {
        location.reload();
      }, 5000);
    };

    tmpWS.onmessage = function(evt: any) {
      const data = JSON.parse(evt.data);
      // console.log(data);
      if (data.MessageType === 0) {
        for (let i = 0; i < globalTorrents.length; i ++) {
          if (globalTorrents[i].HexString === data.HexString) {
            const currentProgress = globalTorrents[i].Percentage;
            if (parseFloat(currentProgress) < parseFloat(data.Percentage)) {
              globalTorrents[i].Percentage = data.Percentage;
              globalTorrents[i].LeftTime = data.LeftTime;
              globalTorrents[i].DownloadSpeed = data.DownloadSpeed;
            }
            if (parseFloat(currentProgress) === parseFloat('1')) {
              globalTorrents[i].Status = 'Completed';
            }
            break;
          }
        }
      } else if (data.MessageType === 1) {
        console.log('Should refresh');
        tmpThis.getTorrents();
        // location.reload();
      }
    };
    tmpWS.onerror = function(evt: Event) {
      console.log('ERROR: ' + evt);
    };
    return tmpWS;
  }

  private getTorrentWebFromData(torrent: Torrent): Torrent {
    torrent.TypeImg = this.aviImg;
    torrent.LeftTime = 'Estimating ...';
    torrent.DownloadSpeed = 'Estimating ...';
    torrent.Interval = -1;
    torrent.StreamURL = this.configService.baseUrl + '/player/' + torrent.HexString;
    return torrent;
  }

  private compareTwoTorrent(a: Torrent, b: Torrent): number {
    let result: boolean;
    if (a.Status !== b.Status) {
      result = a.Status > b.Status;
    } else {
      result = a.TorrentName > b.TorrentName;
    }
    return result ? 1 : 0;
  }

  getTorrents(): void {
    this.torrentService.getSelectedTorrents(this.status)
        .subscribe((datas: Torrent[]) => {
          this.messagesService.add('torrents list update');
          this.torrents = datas;
          globalTorrents = this.torrents;
          for (let i = 0; i < this.torrents.length; i ++) {
            this.torrents[i] = this.getTorrentWebFromData(this.torrents[i]);
            if (this.torrents[i].Status === 'Running') {
              // console.log(this.torrents[i].Status);
              this.updateInfo(this.torrents[i].HexString);
            }
          }
          this.torrents.sort(this.compareTwoTorrent);
        }, error => {
          console.log(error);
        });
  }

  private addOneTorrentService(file: File): void {
    if (_.endsWith(_.lowerCase(file.name), 'torrent')) {
      this.torrentService.addOneTorrent(file)
          .subscribe((IsAdded: boolean) => {
            if (IsAdded) {
              this.messagesService.add('add one torrent successfully');
              this.getTorrents();
            }
          }, error => {
            this.messagesService.add('unable to add this torrent file');
            console.log(error);
          });
    } else {
      alert('Please upload a valid file which ends with .torrent');
    }
  }

  addOneTorrent(files: FileList): void {
    if (files.length !== 0) {
      return this.addOneTorrentService(files[0]);
    }
  }

  judgeEqual(torrentA: Torrent, torrentB: Torrent) {
    if (torrentA !== null && torrentB != null) {
      return torrentA.HexString === torrentB.HexString;
    }
  }

  torrentBlur() {
    this.selectedTorrent = null;
  }

  clickOneTorrent(event: Event, torrent: Torrent) {
    // console.log(torrent);
    event.stopPropagation();
    this.selectedTorrent = torrent;
  }

  rightClick(event: Event, torrent: Torrent) {
    event.stopPropagation();
    this.selectedTorrent = torrent;
    this.rightMenu.popup({
      window: remote.getCurrentWindow()
    });
  }

  private updateInfo(hexString: string) {
    for (let i = 0; i < this.torrents.length; i ++) {
      if (this.torrents[i].HexString === hexString && this.torrents[i].Interval < 0) {
        this.torrents[i].Interval = window.setInterval(this.getInfo, this.torrentService.refleshTime, hexString);
      }
    }
  }

  private getInfo(hexString: string) {
    // console.log('Send Message');
    ws.send(JSON.stringify({
      MessageType: 0,
      HexString: hexString
    }));
  }


  private getTrueFromSelect(torrent: Torrent) {
    if ( torrent === null || torrent === undefined ) {
      return torrent;
    }
    for (let i = 0; i < this.torrents.length; i ++) {
      if (this.torrents[i].HexString === torrent.HexString) {
        return this.torrents[i];
      }
    }
  }

  handleMagnet(content: any) {
    this.modalService.open(content, { centered: true });
  }

  private getTorrentFromInfoHash(infohash: string) {
    return 'https://itorrents.org/torrent/' + infohash + '.torrent';
  }

  downloadMagnet(magnetURL: string) {
    this.modalService.dismissAll();
    magnetURL = _.trim(magnetURL);
    const torrent = magnetDecode(magnetURL);
    if (torrent.infoHash !== undefined && torrent.infoHash !== '' && torrent.infoHash.length === 40) {
      const infoHash = torrent.infoHash.toUpperCase();
      if (currentMagnet === undefined) {
        currentMagnet = magnetURL;
        alert('Add magnet successfully');
        this.webview.downloadURL(this.getTorrentFromInfoHash(infoHash));
        setTimeout(() => {
          if (currentMagnet !== undefined) {
            this.messagesService.add('use engine to resolve magnet');
            currentMagnet = undefined;
            this.sendMagnet(magnetURL);
          } else {
            // console.log('Solve it by itorrents, nothing more');
          }
        }, 10000);
      } else {
        alert('One magnet is handing, please wait a moment');
      }
    } else {
      alert('Invalid infohash');
    }
  }

  startOneTorrent() {
    this.selectedTorrent = this.getTrueFromSelect(this.selectedTorrent);
    if (this.selectedTorrent !== null && this.selectedTorrent !== undefined
        && this.selectedTorrent.Status !== 'Completed' && this.selectedTorrent.Status !== 'Running') {
      const tmpTorrent = this.selectedTorrent;
      this.torrentService.startDownloadOneTorrent(tmpTorrent.HexString)
          .subscribe((data: JSON) => {
            this.getTorrents();
            if (!data['IsDownloading']) {
              this.messagesService.add('Fail to start this task');
            }
          }, error => {
            console.log(error);
          });
    }
  }

  private openInDirectory() {
    this.selectedTorrent = this.getTrueFromSelect(this.selectedTorrent);
    if (this.selectedTorrent !== null && this.selectedTorrent !== undefined) {
      shell.openItem(this.selectedTorrent.StoragePath);
    }
  }

  stopOneTorrent() {
    if (this.selectedTorrent !== null && this.selectedTorrent !== undefined
        && this.selectedTorrent.Status !== 'Completed') {
      this.selectedTorrent = this.getTrueFromSelect(this.selectedTorrent);
      if (this.selectedTorrent.Interval > 0) {
        console.log(this.selectedTorrent.Interval);
        clearInterval(this.selectedTorrent.Interval);
        this.selectedTorrent.Interval = -1;
        if (this.selectedTorrent.Status === 'Running') {
          this.torrentService.stopDownloadOneTorrent(this.selectedTorrent.HexString)
              .subscribe((data: JSON) => {
                this.getTorrents();
                if (!data['IsStopped']) {
                  this.messagesService.add('Fail to stop this task');
                }
              }, error => {
                this.messagesService.add('Error !!!');
                console.log(error);
              });
        }
      }
    }
  }

  delOneTorrent() {
    // console.log(this.selectedTorrent);
    if (this.selectedTorrent !== null && this.selectedTorrent !== undefined) {

      this.selectedTorrent = this.getTrueFromSelect(this.selectedTorrent);
      if (this.selectedTorrent.Interval > 0) {
        clearInterval(this.selectedTorrent.Interval);
        this.selectedTorrent.Interval = -1;
      }

      this.torrentService.deleteDownloadOneTorrent(this.selectedTorrent.HexString)
          .subscribe((data: JSON) => {
            this.getTorrents();
            console.log(data);
            if (!data['IsDeleted']) {
              this.messagesService.add('Failed to delete');
            } else {
              this.messagesService.add('Delete successfully');
            }
            // location.reload();
          }, error => {
            this.messagesService.add('Error !!!');
            console.log(error);
          });
    }
  }

  showInfo(content: any) {
    this.selectedTorrent = this.getTrueFromSelect(this.selectedTorrent);
    if (this.selectedTorrent !== null && this.selectedTorrent !== undefined) {
      // console.log(this.selectedTorrent);
      this.modalService.open(content, {
        centered: true,
        size: 'lg'
      });
    }
  }

  private getBaseHost(): string {
    return _.trimEnd(window.location.href, this.status);
  }

  showPlay() {
    this.selectedTorrent = this.getTrueFromSelect(this.selectedTorrent);
    if (this.selectedTorrent !== null && this.selectedTorrent !== undefined) {
      if (this.selectedTorrent.Status === 'Running') {
        const electronScreen = screen;
        const size = electronScreen.getPrimaryDisplay().workAreaSize;
        const win = new remote.BrowserWindow({
          width: size.width * 0.7,
          height: size.width * 0.7 / 16 * 9,
          autoHideMenuBar: true,
          titleBarStyle: 'hidden',
        });
        const playerUrl = this.getBaseHost() + 'player/' + this.selectedTorrent.HexString;
        this.messagesService.add('Steam torrent now');
        win.loadURL(playerUrl);
        // win.webContents.openDevTools();
      } else {
        alert('Please choose a running task');
      }
    }
  }
}
