import {Component, OnInit} from '@angular/core';
import { Router } from '@angular/router';
import {StorageService} from "../../services/storage.service";

@Component({
  selector: 'app-header',
  templateUrl: './header.component.html',
  styleUrls: ['./header.component.css']
})
export class HeaderComponent implements OnInit{

  constructor(
    private router: Router,
    public storageService: StorageService
  ) { }


  ngOnInit() {
    this.storageService.getRoleFromToken()
  }

  isLoggedIn(): boolean {
    return localStorage.getItem("authToken") != null;
  }

  logout() {
    localStorage.clear();
    this.router.navigate(['/Login']).then();
  }

}
