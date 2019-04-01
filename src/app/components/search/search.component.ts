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
  torlock = require('../../../assets/tools/torlock.png');
  torrentdownload = require('../../../assets/tools/torrentdownload.png');
  constructor() { }

  ngOnInit() {
  }

  openUrl (url: string) {
    shell.openExternal(url);
  }
}
