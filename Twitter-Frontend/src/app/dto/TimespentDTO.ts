export class TimespentDTO {
    tweet_id: string = "";
    timespent: number = 0;
    
    TimespentDTO(tweet_id: string, timespent: number){
        this.tweet_id = tweet_id;
        this.timespent = timespent;
    }


}