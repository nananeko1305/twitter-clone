import { Component, Inject, OnInit } from '@angular/core';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { Router } from '@angular/router';
import { User } from 'src/app/models/user.model';

@Component({
  selector: 'app-follow-component-dialog',
  templateUrl: './follow-component-dialog.component.html',
  styleUrls: ['./follow-component-dialog.component.css']
})
export class FollowComponentDialogComponent implements OnInit {

  constructor(private router: Router,
    public dialogRef: MatDialogRef<FollowComponentDialogComponent>,
    @Inject(MAT_DIALOG_DATA) public data: string[],
    @Inject(MAT_DIALOG_DATA) public type: boolean,
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
