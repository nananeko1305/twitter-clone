import { Component, OnInit } from '@angular/core';
import {Tweet} from "../../models/tweet.model";
import {FormBuilder, FormControl, FormGroup} from "@angular/forms";
import {Search} from "../../models/search";
import {TweetService} from "../../services/tweet.service";
import {MatSnackBar} from "@angular/material/snack-bar";

@Component({
  selector: 'app-search-tweets',
  templateUrl: './search-tweets.component.html',
  styleUrls: ['./search-tweets.component.css']
})
export class SearchTweetsComponent implements OnInit {

  tweets : Tweet[] = []
  formGroup : FormGroup;

  constructor(
    private tweetService: TweetService,
    private _snackBar: MatSnackBar,
    private formBuilder: FormBuilder
  ) {
    this.formGroup = this.formBuilder.group({
      field: ['default'],
      search_str: ['']
    });
  }

  ngOnInit(): void {
  }

  Search() {

    let search : Search = new Search()
    search.search_str = this.formGroup.get('search_str')?.value
    search.field = this.formGroup.get('field')?.value

    if (!search.search_str.includes('#') && search.field == 'hashtag') {
      this.openSnackBar("Please add # at beginning of word when you search tweets by hashtag", "OK")
      return
    }else if (search.search_str.includes(' ') && search.field == 'hashtag') {
      this.openSnackBar("Please remove any space between words when you search tweets by hashtag","OK")
      return
    }

    switch (search.field) {
      case 'hashtag': {
        search.search_type = 'match';
        search.field = 'text'
        break;
      }
      default: {
        search.search_type = 'fuzzy';
        break;
      }
    }

    this.tweetService.SearchTweets(search).subscribe(
      {
        next: (response) => {
          switch (response) {
            case null : {
              this.openSnackBar("We can't find tweets", "OK")
              this.tweets = []
             break;
            }
            default : {
              this.tweets=response
              break;
            }
          }
        },
        error: (error) => {
          console.log(error)
        }
      }
    )
  }

  openSnackBar(message: string, action: string) {
    this._snackBar.open(message, action);
  }

}
