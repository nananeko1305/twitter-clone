import { Component, OnInit } from '@angular/core';
import {ReportDTO} from "../../../dto/reportDTO";
import {ReportService} from "../../../services/reportService";

@Component({
  selector: 'app-report-list',
  templateUrl: './report-list.component.html',
  styleUrls: ['./report-list.component.css']
})
export class ReportListComponent implements OnInit {

  reports: ReportDTO[] = []

  constructor(
    private reportService: ReportService
  ) { }

  ngOnInit(): void {

    this.reportService.Get().subscribe(
      {
        next: (response) => {
          this.reports = response
        }
      }
    )

  }

}
