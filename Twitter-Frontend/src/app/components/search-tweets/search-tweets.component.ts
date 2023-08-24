import { Component, OnInit } from '@angular/core';
import {Tweet} from "../../models/tweet.model";
import {FormControl, FormGroup} from "@angular/forms";
import {Search} from "../../models/search";

@Component({
  selector: 'app-search-tweets',
  templateUrl: './search-tweets.component.html',
  styleUrls: ['./search-tweets.component.css']
})
export class SearchTweetsComponent implements OnInit {

  tweets : Tweet[] = []
  formGroup : FormGroup = new FormGroup({
    field : new FormControl(''),
    search_str : new FormControl('')
  })

  constructor() { }

  ngOnInit(): void {
  }

  Search() {

    let search : Search = new Search()
    search.search_str = this.formGroup.get('search_str')?.value
    search.field = this.formGroup.get('field')?.value


    switch (search.field) {
      case 'hashtag': {
        search.search_type = 'match';
        break;
      }
      default: {
        search.search_type = 'fuzzy';
        break;
      }
    }


    console.log(JSON.stringify(search))
  }

}
