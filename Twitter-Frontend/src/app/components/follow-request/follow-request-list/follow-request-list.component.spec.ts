import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FollowRequestListComponent } from './follow-request-list.component';

describe('FollowRequestListComponent', () => {
  let component: FollowRequestListComponent;
  let fixture: ComponentFixture<FollowRequestListComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ FollowRequestListComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(FollowRequestListComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
