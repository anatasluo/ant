import { Component, OnInit } from '@angular/core';
import { shell } from 'electron';

@Component({
  selector: 'app-about',
  templateUrl: './about.component.html',
  styleUrls: ['./about.component.scss']
})
export class AboutComponent implements OnInit {
  public collapseNum = 1;
  constructor() { }

  ngOnInit() {
  }

  openUrl (url: string) {
    shell.openExternal(url);
  }

}
