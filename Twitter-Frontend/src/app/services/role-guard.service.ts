import { Injectable } from '@angular/core';
import { ActivatedRouteSnapshot, CanActivate, Router, RouterStateSnapshot } from '@angular/router';
import { User } from '../models/user.model';
import { UserService } from './user.service';

@Injectable({
  providedIn: 'root'
})
export class RoleGuard implements CanActivate {

  constructor(private router: Router,
              private userService: UserService) { }

  user: User = new User();
    
  ngOnInit(): void {
    this.userService.GetMe()
      .subscribe({
        next: (data: User) => {
          this.user = data;
        },
        error: (error) => {
          console.log(error);
        }
      })
    
  }

  canActivate(next: ActivatedRouteSnapshot, state: RouterStateSnapshot) {  
      if (this.user.userType == "Regular") {  
          return true;  
      } else { 
      this.router.navigate(['/Main-Page']);  
      return false
    }  
  } 
  
}
