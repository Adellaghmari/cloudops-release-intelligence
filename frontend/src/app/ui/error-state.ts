import { Component, input, output } from '@angular/core';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-error-state',
  imports: [RouterLink],
  template: `
    <section class="error-state" role="alert">
      <p class="eyebrow">Unable to complete this view</p>
      <h1>{{ title() }}</h1>
      <p>{{ message() }}</p>
      @if (code() || requestId()) {
        <p class="technical">
          @if (code()) { <span>{{ code() }}</span> }
          @if (requestId()) { <span>Request {{ requestId() }}</span> }
        </p>
      }
      <div class="actions">
        @if (retryable()) {
          <button type="button" class="button primary" (click)="retry.emit()">Try again</button>
        }
        <a class="button" [routerLink]="backLink()">{{ backLabel() }}</a>
      </div>
    </section>
  `,
})
export class ErrorState {
  readonly title = input.required<string>();
  readonly message = input.required<string>();
  readonly code = input('');
  readonly requestId = input('');
  readonly retryable = input(true);
  readonly backLink = input('/');
  readonly backLabel = input('Back to Operations');
  readonly retry = output<void>();
}
