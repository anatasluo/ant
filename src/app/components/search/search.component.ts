import { Component, OnInit } from '@angular/core';
import { shell } from 'electron';

@Component({
  selector: 'app-search',
  templateUrl: './search.component.html',
  styleUrls: ['./search.component.scss']
})
export class SearchComponent implements OnInit {

  constructor() { }

  ngOnInit() {
  }

  openUrl (url: string) {
    shell.openExternal(url);
  }
}
