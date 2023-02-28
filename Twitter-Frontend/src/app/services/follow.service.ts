import { HttpClient } from '@angular/common/http';
import { Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { environment } from 'src/environments/environment';
import { AdConfig } from '../models/adConfig';
import { FollowRequest } from '../models/followRequest.model';
import { User } from '../models/user.model';

@Injectable({
  providedIn: 'root'
})
export class FollowService {

  private url = "follows"

  constructor(private http: HttpClient) { }

  public GetRequestsForUser(): Observable<any>{
    return this.http.get<any>(`${environment.baseApiUrl}/${this.url}/requests/`)
  }


  public AcceptRequest(id: string): Observable<any> {
    return this.http.put(`${environment.baseApiUrl}/${this.url}/acceptRequest/` + id, null)
  }

  public DeclineRequest(id: string): Observable<any> {
    return this.http.put(`${environment.baseApiUrl}/${this.url}/declineRequest/` + id, null)
  }

  public SendRequest(visibility: string, receiver: FollowRequest): Observable<any>{
    return this.http.post<any>(`${environment.baseApiUrl}/${this.url}/requests/` + visibility, receiver)
  }

  public CreateAdd(adConfig: AdConfig): Observable<void> {
    return this.http.post<any>(`${environment.baseApiUrl}/${this.url}/ad`, adConfig)
  }

  public Recommendations(): Observable<string[]> {
    return this.http.get<string[]>(`${environment.baseApiUrl}/${this.url}/recommendations`)
  }

  public IsFollowExist(username: string): Observable<boolean> {
    return this.http.get<boolean>(`${environment.baseApiUrl}/${this.url}/followExist/${username}`)
  }

  public GetFollowings(username: string): Observable<string[]>{
    return this.http.get<string[]>(`${environment.baseApiUrl}/${this.url}/followings/${username}`)
  }

  public GetFollowiners(username: string): Observable<string[]>{
    return this.http.get<string[]>(`${environment.baseApiUrl}/${this.url}/followers/${username}`)
  }
}
