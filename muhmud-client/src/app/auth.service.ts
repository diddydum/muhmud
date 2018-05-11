import { Injectable } from '@angular/core';
import { HttpClient, HttpErrorResponse, HttpParams, HttpHeaders } from '@angular/common/http';
import { environment } from '../environments/environment';
import { catchError, ignoreElements, map } from 'rxjs/operators';
import { Observable, throwError } from 'rxjs';

interface LoginResult {
  Token: string;
}

@Injectable()
export class AuthService {
  constructor(private http: HttpClient) { }

  get token() {
    return window.localStorage.getItem('token');
  }

  authenticate(email: string, password: string): Observable<void> {
    const body = new HttpParams()
      .set('email', email)
      .set('password', password);
    const response = this.http.post<LoginResult>(environment.server + '/login',
      body.toString(),
      {
        headers: new HttpHeaders()
          .set('Content-Type', 'application/x-www-form-urlencoded')
      });

    // at some point, store the token for later use!
    return response.pipe(
      map(x => {
        window.localStorage.setItem('token', x.Token);
        return void 0;
      }),
      catchError(this.handleError));
  }

  handleError(error: HttpErrorResponse): Observable<Error> {
    if (error.error instanceof ErrorEvent) {
      // client-side or network error occurred
      console.error('An error occurred:', error.error.message);
    } else if (error.status === 0) {
      return throwError(new Error('Unable to connect to server'));
    } else if (error.status === 403) {
      return throwError(new Error('Invalid email/password.'));
    }
    // The backend returned an unsuccessful response code.
    // The response body may contain clues as to what went wrong,
    console.error(
      `Backend returned code ${error.status}, ` +
      `body was: ${error.error}, `);

    return throwError(new Error('Generic failure message, ack'));
  }

  clearToken(): void {
    window.localStorage.removeItem('token');
  }

}
