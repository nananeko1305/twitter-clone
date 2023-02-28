import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { Report } from '../models/report';

@Injectable({
  providedIn: 'root'
})
export class ReportService {

  private url = "reports"

  constructor(private http: HttpClient) { }

  public GetReport(tweet_id: string, timestamp: number, type: string): Observable<Report>{
    return this.http.get<Report>(`${environment.baseApiUrl}/${this.url}/` + tweet_id + "/" + type + "/" + timestamp)
  }
}
