import { HttpClient } from "@angular/common/http";
import { Injectable } from "@angular/core";
import { Observable } from "rxjs";
import { environment } from "src/environments/environment";
import { FeedData } from "../dto/feedData";
import { TimespentDTO } from "../dto/TimespentDTO";
import { TweetID } from "../dto/tweetIdDTO";
import { Favorite } from "../models/favorite.model";
import { Tweet } from "../models/tweet.model";

@Injectable({
    providedIn: 'root'
    })

    export class TweetService {
    private url = "tweets";
    constructor(private http: HttpClient) { }


    public AddTweet(formData: FormData): Observable<Tweet> {
        return this.http.post<Tweet>(`${environment.baseApiUrl}/${this.url}/`, formData);
    }

    public GetHomeFeed(): Observable<FeedData> {
        return this.http.get<FeedData>(`${environment.baseApiUrl}/${this.url}/feed`);
    }

    public GetTweetsForUser(username: string): Observable<Tweet[]> {
        return this.http.get<Tweet[]>(`${environment.baseApiUrl}/${this.url}/user/` + username)
    }

    //Ovde je povezano sa bekom gde fali endpoint
    public GetOneTweetById(tweetID: string): Observable<Tweet> {
        return this.http.get<Tweet>(`${environment.baseApiUrl}/${this.url}/getOneTweet/` + tweetID)
    }

    public LikeTweet(tweet: Tweet): Observable<any> {
        return this.http.post<any>(`${environment.baseApiUrl}/${this.url}/favorite`, tweet)
    }

    public Retweet(tweet: Tweet): Observable<any> {
        return this.http.post<any>(`${environment.baseApiUrl}/${this.url}/retweet`, tweet)
    }

    public GetLikesByTweet(tweetID: string): Observable<Favorite[]> {
        return this.http.get<Favorite[]>(`${environment.baseApiUrl}/${this.url}/whoLiked/` + tweetID)
    }
    
    public GetImageByTweet(tweetID: string): Observable<Blob> {
        return this.http.get(`${environment.baseApiUrl}/${this.url}/image/${tweetID}`, { responseType: 'blob' })
    }

    public TimespentOnAd(timespent: TimespentDTO): Observable<void> {
        return this.http.post<void>(`${environment.baseApiUrl}/${this.url}/timespent`, timespent)
    }

    public ViewedProfileFromAd(tweetID: TweetID): Observable<void> {
        return this.http.post<void>(`${environment.baseApiUrl}/${this.url}/viewCount`, tweetID)
    }

}