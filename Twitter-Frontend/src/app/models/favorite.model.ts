export class Favorite {
    id: string = "";
    username: string = "";
    tweet_id: string = "";

    Tweet(id: string, username: string, tweet_id: string) {
        this.id = id
        this.username = username;
        this.tweet_id = tweet_id;
    }
}