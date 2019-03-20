import { Injectable } from '@angular/core';

@Injectable({
  providedIn: 'root'
})

// TODO: It should load config from file system, just like Viper
export class ConfigService {
  port = '8482';
  host = '127.0.0.1';
  addr = this.host + ':' + this.port;
  baseUrl = 'http://' + this.addr;
  wsBaseUrl = 'ws://' + this.addr + '/ws';

  constructor() { }
}
