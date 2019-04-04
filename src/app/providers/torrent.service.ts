import { Injectable } from '@angular/core';
import { Torrent } from '../classes/torrent';
import { HttpClient, HttpHeaders } from '@angular/common/http';

import { Observable, of} from 'rxjs';
import { catchError, tap} from 'rxjs/operators';

import { ConfigService } from './config.service';

@Injectable({
  providedIn: 'root'
})

export class TorrentService {

  addOneTorrentUrl = this.configService.baseUrl + '/torrent' + '/addOneFile';
  getAllEngineTorrentsUrl = this.configService.baseUrl + '/torrent' + '/getAllEngineTorrents';
  getAllTorrentsUrl = this.configService.baseUrl + '/torrent' + '/getAllTorrents';
  getCompletedTorrentsUrl = this.configService.baseUrl + '/torrent' + '/getCompletedTorrents';
  downloadOneTorrentUrl = this.configService.baseUrl + '/torrent' + '/startDownload';
  stopTorrentUrl = this.configService.baseUrl + '/torrent' + '/stopDownload';
  deleteTorrentUrl = this.configService.baseUrl + '/torrent' + '/delOne';
  sendMagnetUrl = this.configService.baseUrl + '/magnet' + '/addOneMagent';
  refleshTime = 1000;

  private formHttpOptions = {
    headers: new HttpHeaders({
      'enctype': 'multipart/form-data',
      'Access-Control-Allow-Origin': this.configService.addr,
    })
  };

  constructor(
      private httpClient: HttpClient,
      private configService: ConfigService
  ) { }

  addOneTorrent(file: File): Observable<boolean> {
    // console.log(file);
    // console.log(this.baseUrl);
    const formData: FormData = new FormData();
    formData.append('oneTorrentFile', file, file.name);
    return this.httpClient.post<boolean>(this.addOneTorrentUrl, formData, this.formHttpOptions)
        .pipe(
            tap(_ => console.log('addOne torrent')),
            catchError(this.handleError<boolean>('add one torrent'))
        );
  }

  sendMagnet(magnet: string): Observable<boolean> {
    const formData: FormData = new FormData();
    formData.append('linkAddress', magnet);
    return this.httpClient.post<boolean>(this.sendMagnetUrl, formData, this.formHttpOptions)
        .pipe(
            tap(_ => console.log('addOne magnet')),
            catchError(this.handleError<boolean>('add one magnet'))
        );
  }

  private getTorrents(url: string): Observable<Torrent[]> {
    // console.log('url:' + url);
    return this.httpClient.get<Torrent[]>(url)
        .pipe(
            tap(_ => console.log('get torrents')),
            catchError(this.handleError<Torrent[]>('add one torrent'))
        );
  }

  getSelectedTorrents(state: string): Observable<Torrent[]> {
    if (state === 'running') {
      return this.getTorrents(this.getAllEngineTorrentsUrl);
    } else if (state === 'total') {
      return this.getTorrents(this.getAllTorrentsUrl);
    } else if (state === 'completed') {
      return this.getTorrents(this.getCompletedTorrentsUrl);
    }
  }

  startDownloadOneTorrent(hexString: string): Observable<JSON> {
    return this.operateOneTorrent(this.downloadOneTorrentUrl, hexString);
  }

  stopDownloadOneTorrent(hexString: string): Observable<JSON> {
    return this.operateOneTorrent(this.stopTorrentUrl, hexString);
  }

  deleteDownloadOneTorrent(hexString: string): Observable<JSON> {
    return this.operateOneTorrent(this.deleteTorrentUrl, hexString);
  }

  private operateOneTorrent(operateUrl: string, hexString: string): Observable<JSON> {
    const formData: FormData = new FormData();
    formData.append('hexString', hexString);
    return this.httpClient.post<JSON>(operateUrl, formData, this.formHttpOptions)
        .pipe(
            tap(_ => console.log(operateUrl + 'operate torrent')),
            catchError(this.handleError<JSON>('failed to operate one torrent'))
        );
  }

  /**
   * Handle Http operation that failed.
   * Let the app continue.
   * @param operation - name of the operation that failed
   * @param result - optional value to return as the observable result
   */
  private handleError<T>(operation = 'operation', result?: T) {
    return (error: any): Observable<T> => {

      // TODO: send the error to remote logging infrastructure
      console.error(error); // log to console instead

      // Let the app keep running by returning an empty result.
      return of(result as T);
    };
  }

}
