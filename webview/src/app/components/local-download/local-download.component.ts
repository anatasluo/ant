import { Component, OnInit } from '@angular/core';
import { TorrentService } from '../../providers/torrent.service';
import * as _ from 'lodash';
import {Torrent} from '../../classes/torrent';

@Component({
  selector: 'app-local-download',
  templateUrl: './local-download.component.html',
  styleUrls: ['./local-download.component.scss']
})

export class LocalDownloadComponent implements OnInit {

  private aviImg = require('../../../assets/file_type/zip.png');
  torrents: Torrent[];
  selectedTorrent: Torrent;
  constructor(private torrentService: TorrentService) { }

  ngOnInit() {
    this.torrentService.getAllTorrent().subscribe((datas: Torrent[]) => {
      for (let i = 0; i < datas.length; i ++) {
        datas[i].TypeImg = this.aviImg;
      }
      this.torrents = datas;
    }, error => {
      console.log(error);
    });
  }

  addOneTorrent(files: FileList): void {
    // console.log(files.length)
    if (files.length !== 0) {
      if (_.endsWith(_.lowerCase(files[0].name), 'torrent')) {
        const torrentFile: File = files[0];
        // console.log(torrentFile);
        this.torrentService.addOneTorrent(torrentFile).subscribe((data: Torrent) => {
          // TODO
          data.TypeImg = this.aviImg;
          this.torrents.push(data);
        }, error => {
          console.log(error);
        });
      } else {
        alert('请上传有效文件(后缀为.torrent)');
      }

    }
  }

  clickOneTorrent(torrent: Torrent) {
    this.selectedTorrent = torrent;
  }

}
