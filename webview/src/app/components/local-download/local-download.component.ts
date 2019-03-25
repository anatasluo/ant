import {Component, OnInit} from '@angular/core';
import {TorrentService} from '../../providers/torrent.service';
import * as _ from 'lodash';
import {Torrent} from '../../classes/torrent';
import {ActivatedRoute} from '@angular/router';
import { ipcRenderer, remote, screen } from 'electron';
import * as fs from 'fs';
import * as path from 'path';

import { NgbModal } from '@ng-bootstrap/ng-bootstrap';
import {ConfigService} from '../../providers/config.service';

import { magnetDecode } from '@ctrl/magnet-link';

let globalTorrents: Torrent[];
let ws: WebSocket;
let torrentFile: File;

@Component({
  selector: 'app-local-download',
  templateUrl: './local-download.component.html',
  styleUrls: ['./local-download.component.scss']
})

export class LocalDownloadComponent implements OnInit {

  private aviImg = require('../../../assets/file_type/zip.png');
  torrents: Torrent[];
  // See GetTrueFromSelected
  selectedTorrent: Torrent;
  status: string;
  webview: any;
  constructor(private torrentService: TorrentService,
              private configService: ConfigService,
              private route: ActivatedRoute,
              private modalService: NgbModal,
  ) { }
  ngOnInit() {
    this.status = this.route.snapshot.url[0].path;
    this.getTorrents();
    ws = this.generateOneWS();
    this.webview = document.querySelector('webview');
    // @ts-ignore
    this.webview.addEventListener('dom-ready', () => {
      console.log('webview ready!');
    });

    ipcRenderer.on('torrentDownload', (event, arg) => {
      console.log('Download Finished! Handle it now');
      const filePath: string = arg;
      if (_.endsWith(filePath, 'torrent')) {
        this.getFileFromURL(filePath);
      } else {
        console.log('Failed to get meta data!');
      }
    });

    ipcRenderer.on('torrentLoaded', () => {
      this.addOneTorrentService(torrentFile);
    });
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

  private generateOneWS(): WebSocket {
    const tmpWS = new WebSocket(this.configService.wsBaseUrl);
    tmpWS.onopen = function(evt: any) {
      console.log('create websocket');
    };
    tmpWS.onclose = function(evt: any) {
      console.log('close websocket');
      ws.close();
      setTimeout(() => {
        location.reload();
      }, 5000);
    };
    tmpWS.onmessage = function(evt: any) {
      const data = JSON.parse(evt.data);
      // console.log('Get message');
      for (let i = 0; i < globalTorrents.length; i ++) {
        if (globalTorrents[i].HexString === data.HexString) {
          const currentProgress = globalTorrents[i].Percentage;
          if (parseFloat(currentProgress) < parseFloat(data.Percentage)) {
            globalTorrents[i].Percentage = data.Percentage;
            globalTorrents[i].leftTime = data.LeftTime;
            globalTorrents[i].downloadSpeed = data.DownloadSpeed;
          }
          if (parseFloat(currentProgress) === parseFloat('1')) {
            globalTorrents[i].Status = 'Completed';
          }
          break;
        }
      }
    };
    tmpWS.onerror = function(evt: Event) {
      console.log('ERROR: ' + evt);
    };
    return tmpWS;
  }
  private getTorrentWebFromData(torrent: Torrent): Torrent {
    torrent.TypeImg = this.aviImg;
    torrent.leftTime = 'Estimating ...';
    torrent.downloadSpeed = 'Estimating ...';
    torrent.Interval = -1;
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
    this.torrentService.getSelectedTorrents(this.status).subscribe((datas: Torrent[]) => {
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
      console.log(file.name);
      this.torrentService.addOneTorrent(file).subscribe((IsAdded: boolean) => {
        if (IsAdded) {
          this.getTorrents();
        }
      }, error => {
        console.log(error);
      });
    } else {
      alert('请上传有效文件(后缀为.torrent)');
    }
  }

  addOneTorrent(files: FileList): void {
    console.log('Click addOne Torrent');
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
    const torrent = magnetDecode(magnetURL);
    console.log(torrent.infoHash);
    if (torrent.infoHash !== undefined && torrent.infoHash !== '' && torrent.infoHash.length === 40) {
      const infoHash = torrent.infoHash.toUpperCase();
      console.log(infoHash);
      this.webview.downloadURL(this.getTorrentFromInfoHash(infoHash));
    } else {
      console.log('Invalid infohash');
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
              alert('无法下载已完成任务');
            }
          }, error => {
            console.log(error);
          });
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
                  alert('Failed to stop');
                }
              }, error => {
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
              alert('Failed to delete');
            }
            // location.reload();
          }, error => {
            console.log(error);
          });
    }
  }

  showInfo() {
    this.selectedTorrent = this.getTrueFromSelect(this.selectedTorrent);
    if (this.selectedTorrent !== null && this.selectedTorrent !== undefined) {
      console.log(this.selectedTorrent);
    }
  }

  private getBaseHost(): string {
    return _.trimEnd(window.location.href, this.status);
  }

  showPlay() {
    this.selectedTorrent = this.getTrueFromSelect(this.selectedTorrent);
    if (this.selectedTorrent !== null && this.selectedTorrent !== undefined) {
      const electronScreen = screen;
      const size = electronScreen.getPrimaryDisplay().workAreaSize;
      const win = new remote.BrowserWindow({
        width: size.width * 0.7,
        height: size.width * 0.7 / 16 * 9,
        autoHideMenuBar: true,
        titleBarStyle: 'hidden',
      });
      win.loadURL(this.getBaseHost() + 'player/' + this.selectedTorrent.HexString);
      // win.webContents.openDevTools();
    }


  }
}
