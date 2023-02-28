import { Component, OnInit } from '@angular/core';
import { AbstractControl, FormBuilder, FormControl, FormGroup, Validators } from '@angular/forms';
import {Router} from "@angular/router"
import { User } from 'src/app/models/user.model';
import {MatSnackBar} from '@angular/material/snack-bar';
import { AuthService } from 'src/app/services/auth.service';
import { PasswordStrenghtValidator } from 'src/app/services/customValidators';
import { VerificationService } from 'src/app/services/verify.service';
import { HttpErrorResponse } from '@angular/common/http';

@Component({
  selector: 'app-register-regular',
  templateUrl: './register-regular.component.html',
  styleUrls: ['./register-regular.component.css']
})
export class RegisterRegularComponent implements OnInit {

  formGroup: FormGroup = new FormGroup({
    firstName: new FormControl(''),
    lastName: new FormControl(''),
    gender: new FormControl(''), 
    age: new FormControl(''),
    residence: new FormControl(''),
    email: new FormControl(''),
    username: new FormControl(''),
    password: new FormControl('')
  });

  aFormGroup!: FormGroup;
  siteKey: any;

  genders: string[] = [
    'Male',
    'Female'
  ];

  constructor(private authService: AuthService,
              private formBuilder: FormBuilder,
              private router: Router,
              private verificationService: VerificationService,
              private _snackBar: MatSnackBar) { }

  submitted = false;

  ngOnInit(): void {
    this.formGroup = this.formBuilder.group({
      firstName: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(20), Validators.pattern('[-_a-zA-Z]*')]],
      lastName: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(20), Validators.pattern('[-_a-zA-Z]*')]],
      gender: ['', [Validators.required]],
      age: ['', [Validators.required, Validators.min(1), Validators.max(100)]],
      residence: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(35), Validators.pattern('[a-zA-Z ]*')]],//Validators.pattern('[-_a-zA-Z]*')
      email: ['', [Validators.required, Validators.email, Validators.minLength(3), Validators.maxLength(35)]],
      username: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(30), Validators.pattern('[-_a-zA-Z0-9]*')]],
      password: ['', [Validators.required, Validators.minLength(10), Validators.maxLength(30), PasswordStrenghtValidator(), Validators.pattern('[-_a-zA-Z0-9]*')]],
    });

    this.aFormGroup = this.formBuilder.group({
      recaptcha: ['', [Validators.required]]
    });
    this.siteKey = "6LcWR2ojAAAAANOQSFGgbRdboL4Z0xz98_Gpmouz"
  }

  get registerForm(): { [key: string]: AbstractControl } {
    return this.formGroup.controls;
  }

  onSubmit() {
    this.submitted = true;

    if (this.formGroup.invalid) {
      return;
    }

    let registerUser: User = new User();

    registerUser.firstName = this.formGroup.get("firstName")?.value;
    registerUser.lastName = this.formGroup.get("lastName")?.value;
    registerUser.gender = this.formGroup.get("gender")?.value;
    registerUser.age = this.formGroup.get("age")?.value;
    registerUser.residence = this.formGroup.get("residence")?.value;
    registerUser.email = this.formGroup.get("email")?.value;
    registerUser.username = this.formGroup.get("username")?.value;
    registerUser.password = this.formGroup.get("password")?.value;

    this.authService.Register(registerUser)
      .subscribe({
        next: (verificationToken:string) => {
          this.verificationService.updateUserMail(registerUser.email);
          this.verificationService.updateVerificationToken(verificationToken);
          this.router.navigate(['/Verify-Account']);
        },
        error: (error: HttpErrorResponse) => {
          if (error.status == 406) {
            this.openSnackBar(error.error, "Ok")
          }else if (error.status == 302) {
            this.openSnackBar(error.error, "Ok")
          }
        }
      });
  }

  openSnackBar(message: string, action: string) {
    this._snackBar.open(message, action);
  }

}