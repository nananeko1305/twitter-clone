import { Component, OnInit } from '@angular/core';
import {FormControl, FormGroup} from "@angular/forms";
import {Search} from "../../models/search";

@Component({
  selector: 'app-search-people',
  templateUrl: './search-people.component.html',
  styleUrls: ['./search-people.component.css']
})
export class SearchPeopleComponent implements OnInit {

  constructor() { }

  ngOnInit(): void {
  }

  formGroup: FormGroup = new FormGroup({
    searchText: new FormControl('')}
  )

  findPeople() {

    var search: Search = new Search()
    search.search_str = this.formGroup.get('searchText')?.value

    console.log(JSON.stringify(search))

  }


}
