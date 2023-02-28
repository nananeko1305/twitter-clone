import { NgModule } from '@angular/core';
import { BrowserModule } from '@angular/platform-browser';
import { HttpClientModule, HTTP_INTERCEPTORS } from '@angular/common/http';
import { NgxCaptchaModule } from 'ngx-captcha';
import { MatMomentDateModule } from '@angular/material-moment-adapter';
import { MatCardModule } from '@angular/material/card';
import { MatNativeDateModule } from '@angular/material/core';
import { MatButtonModule} from '@angular/material/button';
import { MatMenuModule } from '@angular/material/menu';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatIconModule } from '@angular/material/icon';
import { FormsModule, ReactiveFormsModule } from '@angular/forms';
import {MatFormFieldModule} from '@angular/material/form-field';
import {MatSelectModule} from '@angular/material/select';
import { MatDividerModule } from '@angular/material/divider';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { MainPageComponent } from './components/main-page/main-page.component';
import { HeaderComponent } from './components/header/header.component';
import { TweetViewComponent } from './components/tweet/tweet-view/tweet-view.component';
import { TweetItemComponent } from './components/tweet/tweet-item/tweet-item.component';
import { TweetListComponent } from './components/tweet/tweet-list/tweet-list.component';
import { RegisterRegularComponent } from './components/register-regular/register-regular.component';
import { RegisterBusinessComponent } from './components/register-business/register-business.component';
import { LoginComponent } from './components/login/login.component';
import {MatDatepickerModule} from '@angular/material/datepicker';
import { AuthInterceptor } from './services/auth.interceptor';
import { VerifyAccountComponent } from './components/verify-account/verify-account.component';
import { RecoveryEnterMailComponent } from './components/recovery-enter-mail/recovery-enter-mail.component';
import { RecoveryEnterTokenComponent } from './components/recovery-enter-token/recovery-enter-token.component';
import { RecoveryNewPasswordsComponent } from './components/recovery-new-passwords/recovery-new-passwords.component';
import { TestAuthPageComponent } from './components/test-auth-page/test-auth-page.component';
import { MyProfileComponent } from './components/my-profile/my-profile.component';
import { ChangePasswordComponent } from './components/change-password/change-password.component';
import { TweetAddComponent } from './components/tweet/tweet-add/tweet-add.component';
import { UserProfileComponent } from './components/user-profile/user-profile.component';
import { NotFoundComponent } from './components/not-found/not-found.component';
import { FollowRequestsComponent } from './components/my-follow-requests/follow-requests.component';
import { FollowRequestItemComponent } from './components/follow-request/follow-request-item/follow-request-item.component';
import { FollowRequestListComponent } from './components/follow-request/follow-request-list/follow-request-list.component';
import { TweetLikesDialogComponent } from './components/tweet/tweet-likes-dialog/tweet-likes-dialog.component';
import { MatDialogModule } from '@angular/material/dialog';
import { MatSnackBarModule } from '@angular/material/snack-bar';
import { FollowComponentDialogComponent } from './components/follow-component-dialog/follow-component-dialog.component';
import { FollowingComponentDialogComponent } from './components/following-component-dialog/following-component-dialog.component';
import { MatCarouselModule } from 'ng-mat-carousel';


@NgModule({
  declarations: [
    AppComponent,
    MainPageComponent,
    HeaderComponent,
    TweetViewComponent,
    TweetItemComponent,
    TweetListComponent,
    RegisterRegularComponent,
    RegisterBusinessComponent,
    LoginComponent,
    VerifyAccountComponent,
    RecoveryEnterMailComponent,
    RecoveryEnterTokenComponent,
    RecoveryNewPasswordsComponent,
    TestAuthPageComponent,
    MyProfileComponent,
    ChangePasswordComponent,
    TweetAddComponent,
    UserProfileComponent,
    NotFoundComponent,
    FollowRequestsComponent,
    FollowRequestItemComponent,
    FollowRequestListComponent,
    TweetLikesDialogComponent,
    FollowComponentDialogComponent,
    FollowingComponentDialogComponent
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    HttpClientModule,
    FormsModule,
    ReactiveFormsModule,
    NgxCaptchaModule,
    MatButtonModule,
    MatMenuModule,
    MatToolbarModule,
    MatIconModule,
    MatCardModule,
    MatFormFieldModule,
    MatSelectModule,
    MatDividerModule,
    MatDialogModule,
    BrowserAnimationsModule,
    MatSnackBarModule,
    MatCheckboxModule,
    MatDatepickerModule,
    MatNativeDateModule,
    MatMomentDateModule,
    MatCarouselModule
  ],
  providers: [{
    provide: HTTP_INTERCEPTORS,
    useClass: AuthInterceptor,
    multi: true
  }],
  bootstrap: [AppComponent]
})
export class AppModule { }
