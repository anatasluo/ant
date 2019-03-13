import { Component, OnInit } from '@angular/core';
import { TorrentService } from '../../providers/torrent.service';
import * as _ from "lodash";

@Component({
  selector: 'app-local-download',
  templateUrl: './local-download.component.html',
  styleUrls: ['./local-download.component.css']
})

export class LocalDownloadComponent implements OnInit {

  constructor(private torrentService: TorrentService) { }

  ngOnInit() { }

  addOneTorrent(files: FileList): void {
    // console.log(files.length)
    if (files.length != 0) {
      if (_.endsWith(_.lowerCase(files[0].name), 'torrent')) {
        let torrentFile: File = files[0];
        // console.log(torrentFile);
        this.torrentService.addOneTorrent(torrentFile).subscribe(data => {
          console.log(data);
        }, error => {
          console.log(error);
        });;
      } else {
        alert("请上传有效文件(后缀为.torrent)");
      }

    }
  }

}
