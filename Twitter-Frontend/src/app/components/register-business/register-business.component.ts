import { HttpErrorResponse } from '@angular/common/http';
import { Component, OnInit } from '@angular/core';
import { AbstractControl, FormBuilder, FormControl, FormGroup, ValidationErrors, ValidatorFn, Validators } from '@angular/forms';
import { Router } from '@angular/router';
import { User } from 'src/app/models/user.model';
import { AuthService } from 'src/app/services/auth.service';
import { PasswordStrenghtValidator } from 'src/app/services/customValidators';
import { VerificationService } from 'src/app/services/verify.service';

@Component({
  selector: 'app-register-business',
  templateUrl: './register-business.component.html',
  styleUrls: ['./register-business.component.css']
})
export class RegisterBusinessComponent implements OnInit {

  formGroup: FormGroup = new FormGroup({
    companyName: new FormControl(''),
    email: new FormControl(''),
    website: new FormControl(''), 
    username: new FormControl(''),
    password: new FormControl('')
  });

  aFormGroup!: FormGroup;
  siteKey: any;

  constructor(private authService: AuthService,
              private formBuilder: FormBuilder,
              private router: Router,
              private verificationService: VerificationService) { }

  submitted = false;

  ngOnInit(): void {
    this.formGroup = this.formBuilder.group({
      companyName: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(30), Validators.pattern('[-_a-zA-Z]*')]],
      email: ['', [Validators.required, Validators.email, Validators.minLength(3), Validators.maxLength(35)]],
      website: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(35)]],
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

    registerUser.companyName = this.formGroup.get("companyName")?.value;
    registerUser.email = this.formGroup.get("email")?.value;
    registerUser.website = this.formGroup.get("website")?.value;
    registerUser.username = this.formGroup.get("username")?.value;
    registerUser.password = this.formGroup.get("password")?.value;

    this.authService.Register(registerUser)
      .subscribe({
        next: (verificationToken: string) => {
          this.verificationService.updateUserMail(registerUser.email);
          this.verificationService.updateVerificationToken(verificationToken);
          this.router.navigate(['/Verify-Account']);
        },
        error: (error: HttpErrorResponse) => {
          if (error.status == 406) {
            alert(error.error)
          }
        }
      });
  }

}

export function createPasswordStrenghtValidator(): ValidatorFn {
  return (control: AbstractControl) : ValidationErrors | null => {
    const value = control.value;

    if (!value) {
      return null;
    }

    const hasUpperCase = /[A-Z]+/.test(value)

    const hasLowerCase = /[a-z]+/.test(value);

    const hasNumeric = /[0-9]+/.test(value);

    const passwordValid = hasUpperCase && hasLowerCase && hasNumeric;

    return !passwordValid ? {passwordStrength:true}: null;
  }
}