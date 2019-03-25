import { Component, OnInit } from '@angular/core';
import { ConfigService } from '../../providers/config.service';
import { ActivatedRoute } from '@angular/router';
import { remote } from 'electron';

import { NgbModal } from '@ng-bootstrap/ng-bootstrap';

@Component({
  selector: 'app-player',
  templateUrl: './player.component.html',
  styleUrls: ['./player.component.scss']
})
export class PlayerComponent implements OnInit {

  hexString: string;
  videoUrl: string;
  constructor(
      private configService: ConfigService,
      private route: ActivatedRoute,
      private modalService: NgbModal,
  ) {
  }

  ngOnInit() {
    this.hexString = this.route.snapshot.paramMap.get('hexString');
    this.videoUrl = this.configService.playerUrl + '/' + this.hexString;
    const videoEle = document.getElementsByTagName('video')[0];
    const waringButton = document.getElementById('waringButton');
    waringButton.click();
    videoEle.addEventListener('error', function() {
      alert('Unsupported type');
    }, true);
  }

  handleVideoWaring(content: any) {
    this.modalService.open(content, {centered: true});
  }
}
