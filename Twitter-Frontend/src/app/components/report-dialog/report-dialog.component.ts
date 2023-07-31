import {Component, Inject, OnInit} from '@angular/core';
import {MAT_DIALOG_DATA, MatDialogRef} from "@angular/material/dialog";
import {FormControl, Validators} from "@angular/forms";
import {ReportDTO} from "../../dto/reportDTO";
import {ReportService} from "../../services/reportService";
import {MatSnackBar} from "@angular/material/snack-bar";

export interface DialogData {
  username: string,
  tweetID: string
}
@Component({
  selector: 'app-report-dialog',
  templateUrl: './report-dialog.component.html',
  styleUrls: ['./report-dialog.component.css']
})

export class ReportDialogComponent implements OnInit {

  constructor(
      public dialogRef: MatDialogRef<ReportDialogComponent>,
      @Inject(MAT_DIALOG_DATA) public data: DialogData,
      private reportService: ReportService,
      public snackBar: MatSnackBar
      ) {}

  reasonControl = new FormControl('', [Validators.required]);
  reasons: string[] = [
      "Spam",
      "Hate",
      "InappropriateContent",
      "VerbalAbuse",
      "DontLikeIt",
  ];

  ngOnInit(): void {
  }

  async submitReport() {

      const report: ReportDTO = new ReportDTO()

      if (this.reasonControl.value != null) {
          report.reason = this.reasonControl.value

      }
      report.tweetID = this.data.tweetID
      report.username = this.data.username

      this.reportService.Post(report)
        .subscribe({
          next: () => {
              this.dialogRef.close()
              this.openSnackBar("Tweet successfully reported", "OK")
          },
          error: () => {
            this.openSnackBar("You already reported this Tweet", "OK")
          }
        });

  }

    openSnackBar(message: string, action: string) {
        this.snackBar.open(message, action, {
            duration: 2000,
        });
    }

}

