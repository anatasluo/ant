import { Injectable } from '@angular/core';
import { Torrent } from '../classes/torrent'
import { HttpClient, HttpHeaders } from '@angular/common/http';

import { Observable, of } from 'rxjs';
import { catchError, map, tap } from 'rxjs/operators';

import { ConfigService } from './config.service'

@Injectable({
  providedIn: 'root'
})

export class TorrentService {

  addoneTorrentUrl = this.configService.baseUrl + "/torrent" + "/addOne";

  private formHttpOptions = {
    headers: new HttpHeaders({
      'enctype': 'multipart/form-data',
      'Access-Control-Allow-Origin': this.configService.addr,
    })
  };

  torrents: Torrent[];

  constructor(
    private httpClient: HttpClient,
    private configService: ConfigService
  ) { }

  addOneTorrent(file: File): Observable<any> {
    // console.log(file);
    // console.log(this.baseUrl);
    const formData: FormData = new FormData();
    formData.append('oneTorrentFile', file, file.name);

    return this.httpClient.post(this.addoneTorrentUrl, formData, this.formHttpOptions)
      .pipe(
        catchError(this.handleError<any>('add one torrent'))
      )
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
