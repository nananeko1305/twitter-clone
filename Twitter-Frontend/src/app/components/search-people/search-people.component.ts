import { Component, OnInit } from '@angular/core';
import {FormBuilder, FormControl, FormGroup} from "@angular/forms";
import {Search} from "../../models/search";
import {UserService} from "../../services/user.service";
import {error} from "@angular/compiler-cli/src/transformers/util";
import {User} from "../../models/user.model";
import {MatSnackBar} from "@angular/material/snack-bar";

@Component({
  selector: 'app-search-people',
  templateUrl: './search-people.component.html',
  styleUrls: ['./search-people.component.css']
})
export class SearchPeopleComponent implements OnInit {

  //initialization of variables
  formGroup: FormGroup;
  users: User[];


  constructor(
    private formBuilder: FormBuilder,
    private _snackBar: MatSnackBar,
    private userService: UserService
  ) {
    this.formGroup = this.formBuilder.group({
      field: ['default'],
      search_str: [''],
    });

    this.users = []
  }

  ngOnInit(): void {
  }

  Search() {
    //clear list
    this.users = []
    let search: Search = new Search()
    search.search_str = this.formGroup.get('search_str')?.value
    search.field = this.formGroup.get('field')?.value
    search.search_type = "fuzzy"

    this.userService.Search(search).subscribe({
      next: (response: User[]) => {
        if (response == null) {
          this.openSnackBar("User not exist", "OK")
        }else{
          this.users = response
        }
      },
      error:(err)=>{
        console.log(err)
      }
    })

  }

  openSnackBar(message: string, action: string) {
    this._snackBar.open(message, action);
  }

}
