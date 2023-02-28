export class AdConfig {
    tweet_id:   string = ""
	residence: string = ""
	gender:    string = ""
	age_from:   number = 0 
	age_to:     number = 0

    AdConfig(tweet_id: string, residence: string, gender: string, age_from: number, age_to: number) {
        this.tweet_id = tweet_id
        this.residence = residence;
        this.gender = gender;
        this.age_from = age_from;
        this.age_to = age_to;
    }
}