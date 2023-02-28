import { ComponentFixture, TestBed } from '@angular/core/testing';

import { TestAuthPageComponent } from './test-auth-page.component';

describe('TestAuthPageComponent', () => {
  let component: TestAuthPageComponent;
  let fixture: ComponentFixture<TestAuthPageComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ TestAuthPageComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(TestAuthPageComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
