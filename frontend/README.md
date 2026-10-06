# CloudOps web

Angular 21 release-intelligence console for CloudOps Release Intelligence.

The browser consumes real backend contracts through `/api/v1`; it does not calculate final verdicts or present fixture JSON as API data. Northstar Commerce is always labeled **SYNTHETIC DEMO**. Verified evidence from this repository may be labeled **LIVE PROJECT DATA**.

## Local development

Start the Go API on port 8080, then:

```bash
npm ci
npm start
```

The Angular dev server opens on `http://localhost:4200` and proxies `/api` to `http://localhost:8080`.

## Verification

```bash
npm run lint
npm test
npm run build
npm run e2e
```

Unit tests use Vitest. The Cypress suite covers the recruiter flow and requires a Cypress executable already available in the machine cache. The production bundle is written to `dist/frontend`.
