export class Report {
    tweet_id: string = "";
    timestamp: number = 0;
    like_count: number = 0;
    unlike_count: number = 0;
    view_count: number = 0;
    time_spent: number = 0

    Tweet(tweet_id: string, timestamp: number, like_count: number, unlike_count: number, view_count: number, time_spent: number) {
        this.tweet_id = tweet_id
        this.timestamp = timestamp;
        this.like_count = like_count;
        this.unlike_count = unlike_count;
        this.view_count = view_count;
        this.time_spent = time_spent
    }
}