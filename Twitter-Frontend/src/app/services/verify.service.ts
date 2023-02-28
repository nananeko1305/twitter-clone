import { Injectable } from '@angular/core';
import { BehaviorSubject } from 'rxjs';

@Injectable({
 providedIn: 'root'
})
export class VerificationService {

 private userMail = new BehaviorSubject('');
 currentUserMail = this.userMail.asObservable();

 private verificationToken = new BehaviorSubject('');
 currentVerificationToken = this.verificationToken.asObservable();

 updateVerificationToken(message: string){
   this.verificationToken.next(message)
 }

 updateUserMail(message: string) {
    this.userMail.next(message)
 }
}