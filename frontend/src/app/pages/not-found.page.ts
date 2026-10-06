import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';

@Component({
  selector: 'app-not-found-page',
  imports: [RouterLink],
  template: `
    <section class="error-state">
      <p class="eyebrow">404 · Route not found</p>
      <h1>This evidence surface does not exist</h1>
      <p>Check the URL or return to Operations to continue the release investigation.</p>
      <div class="actions">
        <a class="button primary" routerLink="/">Back to Operations</a>
        <a class="button" routerLink="/releases">Browse Releases</a>
      </div>
    </section>
  `,
})
export class NotFoundPage {}
