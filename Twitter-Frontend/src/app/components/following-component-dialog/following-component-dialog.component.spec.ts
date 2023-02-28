import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FollowingComponentDialogComponent } from './following-component-dialog.component';

describe('FollowingComponentDialogComponent', () => {
  let component: FollowingComponentDialogComponent;
  let fixture: ComponentFixture<FollowingComponentDialogComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ FollowingComponentDialogComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(FollowingComponentDialogComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
