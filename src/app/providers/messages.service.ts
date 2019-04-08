import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})
export class MessagesService {
  private messages: string[] = [];

  constructor() { }

  add(message: string) {
    this.messages.push(message);
  }

  get() {
    let res = '';
    if (this.messages.length !== 0) {
      res = this.messages[0];
      this.messages.shift();
    }
    return res;
  }

  clear() {
    this.messages = [];
  }

}
