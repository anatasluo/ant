import { Component, OnInit } from '@angular/core';
import { shell } from 'electron';

@Component({
  selector: 'app-search',
  templateUrl: './search.component.html',
  styleUrls: ['./search.component.scss']
})
export class SearchComponent implements OnInit {
  itorrents = require('../../../assets/tools/itorrents.png');
  thepiratebay = require('../../../assets/tools/thepiratebay.png');
  torrents = require('../../../assets/tools/torrents.png');
  torrentdownload = require('../../../assets/tools/torrentdownload.png');
  toorgle = require('../../../assets/tools/toorgle.png');
  torrentseeker = require('../../../assets/tools/torrentseeker.png');
  constructor() { }

  ngOnInit() {
  }

  openUrl (url: string) {
    shell.openExternal(url);
  }
}
