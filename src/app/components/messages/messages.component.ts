import { Component, OnInit, OnDestroy } from '@angular/core';
import { MessagesService } from '../../providers/messages.service';

@Component({
  selector: 'app-messages',
  templateUrl: './messages.component.html',
  styleUrls: ['./messages.component.scss']
})
export class MessagesComponent implements OnInit, OnDestroy {

  currentMessage: string;
  private intervalID: number;
  constructor(
      private messageService: MessagesService
  ) { }

  ngOnInit() {
    this.currentMessage = '';
    const tmpThis = this;
    this.intervalID = window.setInterval(() => {
      tmpThis.currentMessage = tmpThis.messageService.get();
    }, 2000);
  }

  ngOnDestroy() {
    clearInterval(this.intervalID);
  }

}
