import { Injectable } from '@angular/core';
import { ConfigService } from './config.service';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import {Settings} from '../classes/settings';

import { Observable, of} from 'rxjs';
import { catchError, tap} from 'rxjs/operators';

@Injectable({
  providedIn: 'root'
})
export class SettingsService {

  getSettingsUrl = this.configService.baseUrl + '/settings' + '/config';
  applySettingsUrl = this.configService.baseUrl + '/settings' + '/apply';

  private formHttpOptions = {
    headers: new HttpHeaders({
      'Content-type': 'application/json; charset=UTF-8',
      'Access-Control-Allow-Origin': this.configService.addr,
    })
  };

  constructor(
      private httpClient: HttpClient,
      private configService: ConfigService
  ) { }

  getSettings(): Observable<Settings> {
    return this.httpClient.get<Settings>(this.getSettingsUrl)
        .pipe(
            tap(_ => console.log('get settings')),
            catchError(this.handleError<Settings>('get settings'))
        );
  }

  applySettings(newSettings: Settings): Observable<boolean> {
    return this.httpClient.post<boolean>(this.applySettingsUrl, newSettings, this.formHttpOptions)
        .pipe(
            tap(_ => console.log('apply settings')),
            catchError(this.handleError<boolean>('apply settings'))
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
