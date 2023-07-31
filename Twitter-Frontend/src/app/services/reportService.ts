import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { Report } from '../models/report';
import {ReportDTO} from "../dto/reportDTO";

@Injectable({
  providedIn: 'root'
})
export class ReportService {

  private url = "reports"

  constructor(private http: HttpClient) { }

  public Get(): Observable<ReportDTO[]>{
    return this.http.get<ReportDTO[]>(`${environment.baseApiUrl}/tweetReports/${this.url}`)
  }

  public Post(report: ReportDTO): Observable<ReportDTO>{
    return this.http.post<ReportDTO>(`${environment.baseApiUrl}/tweetReports/${this.url}`, report)
  }

  public Put(report: ReportDTO): Observable<ReportDTO>{
    return this.http.put<ReportDTO>(`${environment.baseApiUrl}/tweetReports/${this.url}`, report)
  }
}
