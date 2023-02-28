import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { AbstractControl, FormBuilder, FormControl, FormGroup, Validators } from '@angular/forms';
import { MatSnackBar } from '@angular/material/snack-bar';
import { Router } from '@angular/router';
import { LoginDTO } from 'src/app/dto/loginDTO';
import { AuthService } from 'src/app/services/auth.service';
import { VerificationService } from 'src/app/services/verify.service';

@Component({
  selector: 'app-login',
  templateUrl: './login.component.html',
  styleUrls: ['./login.component.css']
})
export class LoginComponent implements OnInit {
  formGroup: FormGroup = new FormGroup({
    username: new FormControl(''),
    password: new FormControl('')
  });

  constructor(
    private authService: AuthService,
    private router: Router,
    private formBuilder: FormBuilder,
    private verificationService: VerificationService,
    private _snackBar: MatSnackBar
  ) { }

  submitted = false;

  ngOnInit(): void {
    this.formGroup = this.formBuilder.group({
      username: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(20)]],
      password: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(20)]],
    });
    this.formGroup.setErrors({ unauthenticated: true})
  }

  get loginGroup(): { [key: string]: AbstractControl } {
    return this.formGroup.controls;
  }

  onSubmit() {
    this.submitted = true;

    if (this.formGroup.invalid) {
      return;
    }

    let login: LoginDTO = new LoginDTO();

    login.username = this.formGroup.get('username')?.value;
    login.password = this.formGroup.get('password')?.value;

    this.authService.Login(login)
      .subscribe({
        next: (token: string) => {
          localStorage.setItem('authToken', token);
          this.router.navigate(['/Main-Page']);
        },
        error: (error: HttpErrorResponse) => {
          if (error.status == 423) {
            let id = error.error.substring(0, error.error.length-1)
            let snackBarMessage = "Your account is locked, because you didn't verify over Email." + " " + "We have sent an email with a token." + " " + "You have been redirected to the verification page."
            this.openSnackBar(snackBarMessage, "Ok")
            this.verificationService.updateVerificationToken(id);
            this.router.navigate(['/Verify-Account']);
            
          }else{
            this.formGroup.setErrors({ unauthenticated: true });
          }
          
        }
      });
  }

  openSnackBar(message: string, action: string) {
    this._snackBar.open(message, action, {duration: 5000});
  }

}
