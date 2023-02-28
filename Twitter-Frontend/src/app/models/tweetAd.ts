import { AddTweetDTO } from "../dto/addTweetDTO";
import { AdConfig } from "./adConfig";

export class TweetAd {
    tweet: AddTweetDTO = new AddTweetDTO()
    adConfig: AdConfig = new AdConfig()

    TweetAd(tweet: AddTweetDTO, adConfig: AdConfig) {
        this.tweet = tweet
        this.adConfig = adConfig;
    }
}