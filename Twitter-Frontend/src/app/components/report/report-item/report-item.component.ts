import {Component, Input, OnInit} from '@angular/core';
import {ReportDTO} from "../../../dto/reportDTO";
import {Tweet} from "../../../models/tweet.model";
import {TweetService} from "../../../services/tweet.service";
import {ReportService} from "../../../services/reportService";
import {_MatSnackBarBase, MatSnackBar} from "@angular/material/snack-bar";

@Component({
  selector: 'app-report-item',
  templateUrl: './report-item.component.html',
  styleUrls: ['./report-item.component.css']
})
export class ReportItemComponent implements OnInit {

  constructor(
    private tweetService: TweetService,
    private reportService: ReportService,
    private snackBar: MatSnackBar
  ) { }

  @Input() report: ReportDTO = new ReportDTO()
  tweet: Tweet = new Tweet()
  ngOnInit(): void {
    this.tweetService.GetOneTweetById(this.report.tweetID).subscribe(
      {
        next:(response) => {
          this.tweet = response
          JSON.stringify(response)
        }
      }
    )
  }

  changeStatus(report: ReportDTO, status: string) {
    this.report.status = status
    this.reportService.Put(report).subscribe(
      {
        next: () =>{
          this.openSnackBar("Report status changed to: " + this.report.status, "OK")
        }
      }
    )
  }

  openSnackBar(message: string, action: string) {
    this.snackBar.open(message, action, {
      duration: 2000,
    });
  }

}
