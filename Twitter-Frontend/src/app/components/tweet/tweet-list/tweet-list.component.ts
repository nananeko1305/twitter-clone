import { Component, Input } from '@angular/core';
import { Tweet } from 'src/app/models/tweet.model';

@Component({
  selector: 'app-tweet-list',
  templateUrl: './tweet-list.component.html',
  styleUrls: ['./tweet-list.component.css']
})
export class TweetListComponent {


  @Input() tweets: Tweet[] = []

}
