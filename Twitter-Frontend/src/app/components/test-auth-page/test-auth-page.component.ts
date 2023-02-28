import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';

@Component({
  selector: 'app-test-auth-page',
  templateUrl: './test-auth-page.component.html',
  styleUrls: ['./test-auth-page.component.css']
})
export class TestAuthPageComponent implements OnInit {

  constructor(private router: Router) { }

  ngOnInit(): void {
    if (localStorage.getItem("authToken") == null) {
      this.router.navigate(["/Login"])
    }
  }

  isLoggedIn(): boolean {
    if (localStorage.getItem("authToken") != null) {
      return true;
    } else {
      return false;
    }
  }

}
