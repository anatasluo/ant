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

  addOneTorrentUrl = this.configService.baseUrl + '/torrent' + '/addOne';
  getAllTorrentUrl = this.configService.baseUrl + '/torrent' + '/getAll';

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

  addOneTorrent(file: File): Observable<Torrent> {
    // console.log(file);
    // console.log(this.baseUrl);
    const formData: FormData = new FormData();
    formData.append('oneTorrentFile', file, file.name);
    return this.httpClient.post<Torrent>(this.addOneTorrentUrl, formData, this.formHttpOptions)
        .pipe(
            tap(_ => console.log('addOne torrent')),
            catchError(this.handleError<Torrent>('add one torrent'))
        );
  }

  getAllTorrent(): Observable<Torrent[]> {
    return this.httpClient.get<Torrent[]>(this.getAllTorrentUrl)
        .pipe(
            tap(_ => console.log('fetched torrents')),
            catchError(this.handleError<Torrent[]>('add one torrent'))
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
