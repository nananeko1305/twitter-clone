import { NgModule } from '@angular/core';
import { AuthGuard } from './services/auth-guard.service';
import { RouterModule, Routes } from '@angular/router';
import { ChangePasswordComponent } from './components/change-password/change-password.component';
import { LoginComponent } from './components/login/login.component';
import { MainPageComponent } from './components/main-page/main-page.component';
import { RecoveryEnterMailComponent } from './components/recovery-enter-mail/recovery-enter-mail.component';
import { RecoveryEnterTokenComponent } from './components/recovery-enter-token/recovery-enter-token.component';
import { MyProfileComponent } from './components/my-profile/my-profile.component';
import { RegisterBusinessComponent } from './components/register-business/register-business.component';
import { RegisterRegularComponent } from './components/register-regular/register-regular.component';
import { TestAuthPageComponent } from './components/test-auth-page/test-auth-page.component';
import { VerifyAccountComponent } from './components/verify-account/verify-account.component';
import { RecoveryNewPasswordsComponent } from './components/recovery-new-passwords/recovery-new-passwords.component';
import { TweetAddComponent } from './components/tweet/tweet-add/tweet-add.component';
import { UserProfileComponent } from './components/user-profile/user-profile.component';
import { NotFoundComponent } from './components/not-found/not-found.component';
import { FollowRequestsComponent } from './components/my-follow-requests/follow-requests.component';
import { TweetViewComponent } from './components/tweet/tweet-view/tweet-view.component';

const routes: Routes = [
  {
    path: "Main-Page",
    component: MainPageComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'Register-Regular',
    component: RegisterRegularComponent
  },
  {
    path: 'Register-Business',
    component: RegisterBusinessComponent
  },
  {
    path: 'Login',
    component: LoginComponent
  },
  {
    path: 'Verify-Account',
    component: VerifyAccountComponent
  },
  {
    path: 'Request-Recovery',
    component: RecoveryEnterMailComponent
  },
  {
    path: 'Recovery-Token',
    component: RecoveryEnterTokenComponent
  },
  {
    path: 'Recovery-Password',
    component: RecoveryNewPasswordsComponent
  },
  {
    path: 'Auth-Test',
    component: TestAuthPageComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'My-Profile',
    component: MyProfileComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'Follow-Requests',
    component: FollowRequestsComponent
  },
  {
    path: 'View-Profile/:username',
    component: UserProfileComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'Change-Password',
    component: ChangePasswordComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'New-Tweet',
    component: TweetAddComponent,
    canActivate: [AuthGuard]
  },
  {
    path: 'View-Tweet/:id',
    component: TweetViewComponent,
    canActivate: [AuthGuard]
  },
  {
    path: '404',
    component: NotFoundComponent
  },
  {
    path: '**',
    component: NotFoundComponent
  }
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
