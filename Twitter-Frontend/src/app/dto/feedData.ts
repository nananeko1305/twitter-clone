import { Tweet } from "../models/tweet.model";

export class FeedData {
    feed: Tweet[] = [];
    ads: Tweet[] = [];

    ChangePasswordDTO(feed: Tweet[], ads: Tweet[]) {
        this.feed = feed;
        this.ads = ads;
    }
}