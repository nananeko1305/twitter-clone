export class AddTweetDTO {
    text: string = "";
    advertisement: boolean = true

    Tweet(text: string, advertisement: boolean) {
        this.text = text;
        this.advertisement = advertisement
    }
}