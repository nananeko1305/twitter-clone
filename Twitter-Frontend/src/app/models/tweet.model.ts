export class Tweet {
    id: string = "";
    text: string = "";
    created_on: number = 0;
    username: string = "";
    favorite_count: number = 0;
    image: boolean = false;
    advertisement: boolean = false;
    owner_username: string = "";

    Tweet(id: string, text: string, created_on: number, username: string, favorite_count: number, image: boolean, advertisement: boolean, owner_username: string) {
        this.id = id
        this.created_on = created_on
        this.text = text;
        this.username = username;
        this.favorite_count = favorite_count;
        this.image = image;
        this.advertisement = advertisement;
        this.owner_username = owner_username;
    }
}