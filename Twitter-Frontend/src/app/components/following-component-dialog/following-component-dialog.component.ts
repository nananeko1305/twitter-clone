import { Component, Inject, OnInit } from '@angular/core';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { Router } from '@angular/router';
import { User } from 'src/app/models/user.model';

@Component({
  selector: 'app-following-component-dialog',
  templateUrl: './following-component-dialog.component.html',
  styleUrls: ['./following-component-dialog.component.css']
})
export class FollowingComponentDialogComponent implements OnInit {

  constructor(private router: Router,
    public dialogRef: MatDialogRef<FollowingComponentDialogComponent>,
    @Inject(MAT_DIALOG_DATA) public data: string[],
    ) { }

  ngOnInit(): void {
  }

  onOkClick(): void {
    this.dialogRef.close();
  }

  onUsernameClick(username: string): void {
    this.router.navigate(["/View-Profile/" + username])
    this.dialogRef.close("username");
  }

}
