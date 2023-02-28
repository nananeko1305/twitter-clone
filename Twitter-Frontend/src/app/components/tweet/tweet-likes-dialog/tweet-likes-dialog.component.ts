import { Component, Inject } from '@angular/core';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { Router } from '@angular/router';
import { Favorite } from 'src/app/models/favorite.model';
import { User } from 'src/app/models/user.model';

export interface DialogData {
  testUsers: User[]
}

@Component({
  selector: 'app-tweet-likes-dialog',
  templateUrl: './tweet-likes-dialog.component.html',
  styleUrls: ['./tweet-likes-dialog.component.css']
})
export class TweetLikesDialogComponent {

  constructor(
    private router: Router,
    public dialogRef: MatDialogRef<TweetLikesDialogComponent>,
    @Inject(MAT_DIALOG_DATA) public data: Favorite[],
  ) {}

  onOkClick(): void {
    this.dialogRef.close();
  }

  onUsernameClick(username: string): void {
    this.router.navigate(["/View-Profile/" + username])
    this.dialogRef.close("username");
  }

}
